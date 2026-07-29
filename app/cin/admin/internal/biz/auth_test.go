package biz

import (
	"context"
	"io"
	"testing"

	server "cyi-app/pkg/kratos/server/middleware"

	kratoslog "github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// mockAuthRepo 模拟 AuthRepo 接口
type mockAuthRepo struct {
	verifyLoginFn func(ctx context.Context, username, password string) (*AdminAuth, error)
	saveTokenFn   func(ctx context.Context, userID int64, token string) error
	removeTokenFn func(ctx context.Context, token string) error
	getAdminInfoFn func(ctx context.Context, adminID int64) (*AdminAuth, error)
	updateAvatarFn func(ctx context.Context, adminID int64, avatarBase64 string) (string, error)
	updateAdminInfoFn func(ctx context.Context, adminID int64, info *AdminAuth) error
	getOnlineAdminIdsFn func(ctx context.Context) ([]int64, error)
}

func (m *mockAuthRepo) VerifyLogin(ctx context.Context, username, password string) (*AdminAuth, error) {
	return m.verifyLoginFn(ctx, username, password)
}

func (m *mockAuthRepo) SaveToken(ctx context.Context, userID int64, token string) error {
	return m.saveTokenFn(ctx, userID, token)
}

func (m *mockAuthRepo) RemoveToken(ctx context.Context, token string) error {
	return m.removeTokenFn(ctx, token)
}

func (m *mockAuthRepo) GetAdminInfo(ctx context.Context, adminID int64) (*AdminAuth, error) {
	return m.getAdminInfoFn(ctx, adminID)
}

func (m *mockAuthRepo) UpdateAvatar(ctx context.Context, adminID int64, avatarBase64 string) (string, error) {
	return m.updateAvatarFn(ctx, adminID, avatarBase64)
}

func (m *mockAuthRepo) UpdateAdminInfo(ctx context.Context, adminID int64, info *AdminAuth) error {
	return m.updateAdminInfoFn(ctx, adminID, info)
}

func (m *mockAuthRepo) GetOnlineAdminIds(ctx context.Context) ([]int64, error) {
	return m.getOnlineAdminIdsFn(ctx)
}

// TestAuthUseCase_Login_Success 测试登录成功：JWT token 生成并包含正确的 claims
func TestAuthUseCase_Login_Success(t *testing.T) {
	apiKey := "test-api-key-12345"
	expectedAdmin := &AdminAuth{
		ID:       1,
		Username: "admin",
		Name:     "管理员",
		Role:     "super_admin",
	}

	uc := NewAuthUseCase(&mockAuthRepo{
		verifyLoginFn: func(_ context.Context, username, password string) (*AdminAuth, error) {
			assert.Equal(t, "admin", username)
			assert.Equal(t, "password123", password)
			return expectedAdmin, nil
		},
		saveTokenFn: func(_ context.Context, userID int64, token string) error {
			assert.Equal(t, int64(1), userID)
			assert.NotEmpty(t, token)
			return nil
		},
	}, kratoslog.NewStdLogger(io.Discard), apiKey)

	admin, token, err := uc.Login(context.Background(), "admin", "password123")
	assert.NoError(t, err)
	assert.NotNil(t, admin)
	assert.NotEmpty(t, token)
	assert.Equal(t, expectedAdmin.ID, admin.ID)
	assert.Equal(t, expectedAdmin.Username, admin.Username)

	// 验证 JWT token 的 claims
	claims := &server.UserClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(apiKey), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)
	assert.Equal(t, int64(1), claims.ID)
	assert.Equal(t, "admin", claims.Username)
	assert.Equal(t, "super_admin", claims.Role)
	assert.Equal(t, "cin.admin", claims.Issuer)
}

// TestAuthUseCase_Login_VerifyLoginError 测试登录失败时错误传播
func TestAuthUseCase_Login_VerifyLoginError(t *testing.T) {
	expectedErr := io.ErrUnexpectedEOF
	uc := NewAuthUseCase(&mockAuthRepo{
		verifyLoginFn: func(_ context.Context, username, password string) (*AdminAuth, error) {
			return nil, expectedErr
		},
	}, kratoslog.NewStdLogger(io.Discard), "test-key")

	admin, token, err := uc.Login(context.Background(), "admin", "wrong")
	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, admin)
	assert.Empty(t, token)
}

// TestAuthUseCase_Login_SaveTokenError 测试登录验证成功后保存 token 失败时错误传播
func TestAuthUseCase_Login_SaveTokenError(t *testing.T) {
	expectedErr := io.ErrClosedPipe
	uc := NewAuthUseCase(&mockAuthRepo{
		verifyLoginFn: func(_ context.Context, username, password string) (*AdminAuth, error) {
			return &AdminAuth{ID: 1, Username: "admin", Role: "admin"}, nil
		},
		saveTokenFn: func(_ context.Context, userID int64, token string) error {
			return expectedErr
		},
	}, kratoslog.NewStdLogger(io.Discard), "test-key")

	admin, token, err := uc.Login(context.Background(), "admin", "password")
	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, admin)
	assert.Empty(t, token)
}

// TestAuthUseCase_Login_JWTClaimsFormat 测试 JWT claims 格式正确（包含必要字段）
func TestAuthUseCase_Login_JWTClaimsFormat(t *testing.T) {
	apiKey := "jwt-claims-test-key"
	uc := NewAuthUseCase(&mockAuthRepo{
		verifyLoginFn: func(_ context.Context, username, password string) (*AdminAuth, error) {
			return &AdminAuth{ID: 42, Username: "testuser", Role: "editor"}, nil
		},
		saveTokenFn: func(_ context.Context, userID int64, token string) error {
			return nil
		},
	}, kratoslog.NewStdLogger(io.Discard), apiKey)

	_, token, err := uc.Login(context.Background(), "testuser", "pass")
	assert.NoError(t, err)

	// 验证 JWT claims 结构
	claims := &server.UserClaims{}
	parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(apiKey), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// 验证必填字段
	assert.Equal(t, int64(42), claims.ID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "editor", claims.Role)
	assert.Equal(t, "cin.admin", claims.Issuer)

	// 验证时间字段存在（ExpiresAt, IssuedAt, NotBefore）
	assert.NotNil(t, claims.ExpiresAt)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.NotBefore)
	// JWT ID 应为非空 UUID
	assert.NotEmpty(t, claims.ID)
}
