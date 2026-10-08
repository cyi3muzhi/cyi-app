package data

import (
	"context"
	"testing"
	"time"

	"cyi-app/app/service/user/internal/biz"
	"cyi-app/app/service/user/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Skip("SQLite requires CGO, skipping DB-related tests")
	}
	err = db.AutoMigrate(&models.AdminUser{})
	assert.NoError(t, err)
	return db
}

func setupTestRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := client.Ping(ctx).Err()
	if err != nil {
		t.Skip("Redis not available, skipping Redis-related tests")
	}
	return client
}

func TestAdminUserRepo_FindByUsername_Success(t *testing.T) {
	db := setupTestDB(t)

	admin := &models.AdminUser{
		Username: "admin",
		Password: "hashed",
		Name:     "管理员",
		Role:     "超级管理员",
		Status:   1,
	}
	db.Create(admin)

	repo := NewAdminUserRepo(&Data{db: db}, log.NewStdLogger(nil))
	found, err := repo.FindByUsername(context.Background(), "admin")

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, int64(1), found.ID)
	assert.Equal(t, "admin", found.Username)
	assert.Equal(t, "管理员", found.Name)
}

func TestAdminUserRepo_FindByUsername_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewAdminUserRepo(&Data{db: db}, log.NewStdLogger(nil))
	found, err := repo.FindByUsername(context.Background(), "notexist")

	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestAdminUserRepo_FindByID_Success(t *testing.T) {
	db := setupTestDB(t)

	admin := &models.AdminUser{
		Username: "admin",
		Password: "hashed",
		Name:     "管理员",
	}
	db.Create(admin)

	repo := NewAdminUserRepo(&Data{db: db}, log.NewStdLogger(nil))
	found, err := repo.FindByID(context.Background(), admin.ID)

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, admin.ID, found.ID)
	assert.Equal(t, "admin", found.Username)
}

func TestAdminUserRepo_FindByID_NotFound(t *testing.T) {
	db := setupTestDB(t)

	repo := NewAdminUserRepo(&Data{db: db}, log.NewStdLogger(nil))
	found, err := repo.FindByID(context.Background(), 999)

	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestAdminUserRepo_Create_Success(t *testing.T) {
	db := setupTestDB(t)

	repo := NewAdminUserRepo(&Data{db: db}, log.NewStdLogger(nil))
	id, err := repo.Create(context.Background(), &biz.AdminUser{
		Username: "test",
		Password: "hashed",
		Name:     "测试",
		Role:     "管理员",
		Email:    "test@test.com",
		Status:   1,
	})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), id)

	var admin models.AdminUser
	db.First(&admin, id)
	assert.Equal(t, "test", admin.Username)
	assert.Equal(t, "测试", admin.Name)
}

func TestAdminUserRepo_Update_Success(t *testing.T) {
	db := setupTestDB(t)

	admin := &models.AdminUser{
		Username: "admin",
		Password: "hashed",
		Name:     "管理员",
		Role:     "管理员",
		Status:   1,
	}
	db.Create(admin)

	repo := NewAdminUserRepo(&Data{db: db}, log.NewStdLogger(nil))
	err := repo.Update(context.Background(), &biz.AdminUser{
		ID:    admin.ID,
		Name:  "新名称",
		Role:  "超级管理员",
		Email: "new@test.com",
	})

	assert.NoError(t, err)

	db.First(&admin, admin.ID)
	assert.Equal(t, "新名称", admin.Name)
	assert.Equal(t, "超级管理员", admin.Role)
	assert.Equal(t, "new@test.com", admin.Email)
}

func TestAdminUserRepo_Update_NoChanges(t *testing.T) {
	db := setupTestDB(t)

	admin := &models.AdminUser{
		Username: "admin",
		Password: "hashed",
		Name:     "管理员",
	}
	db.Create(admin)

	repo := NewAdminUserRepo(&Data{db: db}, log.NewStdLogger(nil))
	err := repo.Update(context.Background(), &biz.AdminUser{ID: admin.ID})

	assert.NoError(t, err)
}

func TestAdminUserRepo_ToggleStatus_Success(t *testing.T) {
	db := setupTestDB(t)

	admin := &models.AdminUser{
		Username: "admin",
		Password: "hashed",
		Status:   1,
	}
	db.Create(admin)

	repo := NewAdminUserRepo(&Data{db: db}, log.NewStdLogger(nil))
	newStatus, err := repo.ToggleStatus(context.Background(), admin.ID)

	assert.NoError(t, err)
	assert.Equal(t, int32(2), newStatus)

	db.First(&admin, admin.ID)
	assert.Equal(t, int32(2), admin.Status)
}

func TestAdminUserRepo_ToggleStatus_Again(t *testing.T) {
	db := setupTestDB(t)

	admin := &models.AdminUser{
		Username: "admin",
		Password: "hashed",
		Status:   2,
	}
	db.Create(admin)

	repo := NewAdminUserRepo(&Data{db: db}, log.NewStdLogger(nil))
	newStatus, err := repo.ToggleStatus(context.Background(), admin.ID)

	assert.NoError(t, err)
	assert.Equal(t, int32(1), newStatus)
}

func TestParseJWTClaims_Success(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  float64(123),
		"jti": "test-jti-123",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iss": "cin",
	})
	signedToken, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	adminID, jti, err := parseJWTClaims(signedToken)

	assert.NoError(t, err)
	assert.Equal(t, int64(123), adminID)
	assert.Equal(t, "test-jti-123", jti)
}

func TestParseJWTClaims_MissingID(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti": "test-jti",
	})
	signedToken, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	_, _, err = parseJWTClaims(signedToken)

	assert.Error(t, err)
}

func TestParseJWTClaims_MissingJTI(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": float64(123),
	})
	signedToken, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	_, _, err = parseJWTClaims(signedToken)

	assert.Error(t, err)
}

func TestParseJWTClaims_InvalidToken(t *testing.T) {
	_, _, err := parseJWTClaims("invalid-token-string")

	assert.Error(t, err)
}

func TestParseJWTClaims_InvalidClaims(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{})
	signedToken, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	_, _, err = parseJWTClaims(signedToken)

	assert.Error(t, err)
}

func TestSaveToken_Success(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	defer rdb.FlushDB(context.Background())

	repo := NewAdminUserRepo(&Data{db: db, cache: rdb}, log.NewStdLogger(nil))
	err := repo.SaveToken(context.Background(), "test-token", 1)

	assert.NoError(t, err)

	stored, err := rdb.Get(context.Background(), "admin_token:1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "test-token", stored)
}

func TestRemoveToken_Success(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	defer rdb.FlushDB(context.Background())

	rdb.Set(context.Background(), "admin_token:1", "test-token", 24*time.Hour)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  float64(1),
		"jti": "test-jti",
	})
	signedToken, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	repo := NewAdminUserRepo(&Data{db: db, cache: rdb}, log.NewStdLogger(nil))
	err = repo.RemoveToken(context.Background(), signedToken)

	assert.NoError(t, err)

	_, err = rdb.Get(context.Background(), "admin_token:1").Result()
	assert.Error(t, err)
	assert.Equal(t, redis.Nil, err)
}

func TestValidateToken_Valid(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	defer rdb.FlushDB(context.Background())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  float64(1),
		"jti": "test-jti",
	})
	signedToken, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	rdb.Set(context.Background(), "admin_token:1", signedToken, 24*time.Hour)

	repo := NewAdminUserRepo(&Data{db: db, cache: rdb}, log.NewStdLogger(nil))
	valid, adminID, username, err := repo.ValidateToken(context.Background(), signedToken)

	assert.NoError(t, err)
	assert.True(t, valid)
	assert.Equal(t, int64(1), adminID)
	assert.Equal(t, "", username)
}

func TestValidateToken_NotInRedis(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	defer rdb.FlushDB(context.Background())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  float64(1),
		"jti": "test-jti",
	})
	signedToken, err := token.SignedString([]byte("secret"))
	assert.NoError(t, err)

	repo := NewAdminUserRepo(&Data{db: db, cache: rdb}, log.NewStdLogger(nil))
	valid, adminID, username, err := repo.ValidateToken(context.Background(), signedToken)

	assert.NoError(t, err)
	assert.False(t, valid)
	assert.Equal(t, int64(1), adminID)
	assert.Equal(t, "", username)
}

func TestValidateToken_TokenOverwritten(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	defer rdb.FlushDB(context.Background())

	token1 := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  float64(1),
		"jti": "test-jti-1",
	})
	signedToken1, err := token1.SignedString([]byte("secret"))
	assert.NoError(t, err)

	token2 := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  float64(1),
		"jti": "test-jti-2",
	})
	signedToken2, err := token2.SignedString([]byte("secret"))
	assert.NoError(t, err)

	rdb.Set(context.Background(), "admin_token:1", signedToken2, 24*time.Hour)

	repo := NewAdminUserRepo(&Data{db: db, cache: rdb}, log.NewStdLogger(nil))
	valid, adminID, username, err := repo.ValidateToken(context.Background(), signedToken1)

	assert.NoError(t, err)
	assert.False(t, valid)
	assert.Equal(t, int64(1), adminID)
	assert.Equal(t, "", username)
}

func TestValidateToken_InvalidJWT(t *testing.T) {
	db := setupTestDB(t)
	rdb := setupTestRedis(t)
	defer rdb.FlushDB(context.Background())

	repo := NewAdminUserRepo(&Data{db: db, cache: rdb}, log.NewStdLogger(nil))
	_, _, _, err := repo.ValidateToken(context.Background(), "invalid-jwt-token")

	assert.Error(t, err)
}

func TestModelToBiz_Nil(t *testing.T) {
	result := modelToBiz(nil)
	assert.Nil(t, result)
}

func TestModelToBiz_Success(t *testing.T) {
	model := &models.AdminUser{
		Username:  "admin",
		Password:  "hashed",
		Name:      "管理员",
		Role:      "超级管理员",
		Email:     "admin@test.com",
		Status:    1,
		LastLogin: time.Now().Unix(),
	}
	model.ID = 1
	model.CreateTime = time.Now().Unix()

	result := modelToBiz(model)

	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.ID)
	assert.Equal(t, "admin", result.Username)
	assert.Equal(t, "管理员", result.Name)
	assert.Equal(t, int32(1), result.Status)
}

func TestBizToModel_Success(t *testing.T) {
	biz := &biz.AdminUser{
		Username: "admin",
		Password: "hashed",
		Name:     "管理员",
		Role:     "超级管理员",
		Email:    "admin@test.com",
		Status:   1,
	}

	result := bizToModel(biz)

	assert.NotNil(t, result)
	assert.Equal(t, "admin", result.Username)
	assert.Equal(t, "管理员", result.Name)
	assert.Equal(t, int32(1), result.Status)
}
