package biz

import (
	"context"
	"errors"
	"testing"

	v1 "cyi-app/api/service/user/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type mockAdminUserRepo struct {
	findByUsernameFn    func(ctx context.Context, username string) (*AdminUser, error)
	findByIDFn          func(ctx context.Context, id int64) (*AdminUser, error)
	createFn            func(ctx context.Context, admin *AdminUser) (int64, error)
	updateFn            func(ctx context.Context, admin *AdminUser) error
	listFn              func(ctx context.Context, page, pageSize int32, keyword, status string) ([]*AdminUser, int64, error)
	toggleStatusFn      func(ctx context.Context, id int64) (int32, error)
	updateLastLoginFn   func(ctx context.Context, id int64) error
	verifyPasswordFn    func(ctx context.Context, id int64, password string) error
	saveTokenFn         func(ctx context.Context, token string, adminID int64) error
	removeTokenFn       func(ctx context.Context, token string) error
	validateTokenFn     func(ctx context.Context, token string) (bool, int64, string, error)
}

func (m *mockAdminUserRepo) FindByUsername(ctx context.Context, username string) (*AdminUser, error) {
	if m.findByUsernameFn != nil {
		return m.findByUsernameFn(ctx, username)
	}
	return nil, nil
}

func (m *mockAdminUserRepo) FindByID(ctx context.Context, id int64) (*AdminUser, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockAdminUserRepo) Create(ctx context.Context, admin *AdminUser) (int64, error) {
	if m.createFn != nil {
		return m.createFn(ctx, admin)
	}
	return 0, nil
}

func (m *mockAdminUserRepo) Update(ctx context.Context, admin *AdminUser) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, admin)
	}
	return nil
}

func (m *mockAdminUserRepo) List(ctx context.Context, page, pageSize int32, keyword, status string) ([]*AdminUser, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, page, pageSize, keyword, status)
	}
	return nil, 0, nil
}

func (m *mockAdminUserRepo) ToggleStatus(ctx context.Context, id int64) (int32, error) {
	if m.toggleStatusFn != nil {
		return m.toggleStatusFn(ctx, id)
	}
	return 0, nil
}

func (m *mockAdminUserRepo) UpdateLastLogin(ctx context.Context, id int64) error {
	if m.updateLastLoginFn != nil {
		return m.updateLastLoginFn(ctx, id)
	}
	return nil
}

func (m *mockAdminUserRepo) VerifyPassword(ctx context.Context, id int64, password string) error {
	if m.verifyPasswordFn != nil {
		return m.verifyPasswordFn(ctx, id, password)
	}
	return nil
}

func (m *mockAdminUserRepo) SaveToken(ctx context.Context, token string, adminID int64) error {
	if m.saveTokenFn != nil {
		return m.saveTokenFn(ctx, token, adminID)
	}
	return nil
}

func (m *mockAdminUserRepo) RemoveToken(ctx context.Context, token string) error {
	if m.removeTokenFn != nil {
		return m.removeTokenFn(ctx, token)
	}
	return nil
}

func (m *mockAdminUserRepo) ValidateToken(ctx context.Context, token string) (bool, int64, string, error) {
	if m.validateTokenFn != nil {
		return m.validateTokenFn(ctx, token)
	}
	return false, 0, "", nil
}

func TestAdminUserUseCase_Login_Success(t *testing.T) {
	repo := &mockAdminUserRepo{
		findByUsernameFn: func(ctx context.Context, username string) (*AdminUser, error) {
			if username == "admin" {
				return &AdminUser{
					ID:       1,
					Username: "admin",
					Name:     "管理员",
					Role:     "超级管理员",
					Status:   1,
				}, nil
			}
			return nil, nil
		},
		verifyPasswordFn: func(ctx context.Context, id int64, password string) error {
			if password == "admin123" {
				return nil
			}
			return errors.New("密码错误")
		},
		updateLastLoginFn: func(ctx context.Context, id int64) error {
			return nil
		},
		saveTokenFn: func(ctx context.Context, token string, adminID int64) error {
			return nil
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	admin, token, err := uc.Login(context.Background(), "admin", "admin123")

	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if admin == nil {
		t.Fatal("管理员信息为空")
	}
	if admin.ID != 1 {
		t.Errorf("期望ID=1，实际=%d", admin.ID)
	}
	if admin.Username != "admin" {
		t.Errorf("期望用户名=admin，实际=%s", admin.Username)
	}
	if token == "" {
		t.Error("token为空")
	}
}

func TestAdminUserUseCase_Login_UserNotFound(t *testing.T) {
	repo := &mockAdminUserRepo{
		findByUsernameFn: func(ctx context.Context, username string) (*AdminUser, error) {
			return nil, nil
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	_, _, err := uc.Login(context.Background(), "notexist", "password")

	if err == nil {
		t.Fatal("期望错误，实际没有")
	}
	if !errors.Is(err, v1.ErrorUserNotFound("")) {
		t.Errorf("期望 UserNotFound 错误，实际=%v", err)
	}
}

func TestAdminUserUseCase_Login_UserFrozen(t *testing.T) {
	repo := &mockAdminUserRepo{
		findByUsernameFn: func(ctx context.Context, username string) (*AdminUser, error) {
			return &AdminUser{
				ID:       1,
				Username: "admin",
				Status:   2,
			}, nil
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	_, _, err := uc.Login(context.Background(), "admin", "password")

	if err == nil {
		t.Fatal("期望错误，实际没有")
	}
	if !errors.Is(err, v1.ErrorUserNotFound("")) {
		t.Errorf("期望用户已冻结错误，实际=%v", err)
	}
}

func TestAdminUserUseCase_Login_WrongPassword(t *testing.T) {
	repo := &mockAdminUserRepo{
		findByUsernameFn: func(ctx context.Context, username string) (*AdminUser, error) {
			return &AdminUser{
				ID:       1,
				Username: "admin",
				Status:   1,
			}, nil
		},
		verifyPasswordFn: func(ctx context.Context, id int64, password string) error {
			return errors.New("密码错误")
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	_, _, err := uc.Login(context.Background(), "admin", "wrongpassword")

	if err == nil {
		t.Fatal("期望错误，实际没有")
	}
	if !errors.Is(err, v1.ErrorPasswordError("")) {
		t.Errorf("期望密码错误，实际=%v", err)
	}
}

func TestAdminUserUseCase_Logout_Success(t *testing.T) {
	repo := &mockAdminUserRepo{
		removeTokenFn: func(ctx context.Context, token string) error {
			return nil
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	err := uc.Logout(context.Background(), "test-token")

	if err != nil {
		t.Errorf("登出失败: %v", err)
	}
}

func TestAdminUserUseCase_Logout_Error(t *testing.T) {
	expectedErr := errors.New("Redis连接失败")
	repo := &mockAdminUserRepo{
		removeTokenFn: func(ctx context.Context, token string) error {
			return expectedErr
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	err := uc.Logout(context.Background(), "test-token")

	if err != expectedErr {
		t.Errorf("期望错误=%v，实际=%v", expectedErr, err)
	}
}

func TestAdminUserUseCase_ValidateToken_Valid(t *testing.T) {
	repo := &mockAdminUserRepo{
		validateTokenFn: func(ctx context.Context, token string) (bool, int64, string, error) {
			return true, 1, "admin", nil
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	valid, adminID, username, err := uc.ValidateToken(context.Background(), "valid-token")

	if err != nil {
		t.Errorf("验证失败: %v", err)
	}
	if !valid {
		t.Error("期望有效，实际无效")
	}
	if adminID != 1 {
		t.Errorf("期望adminID=1，实际=%d", adminID)
	}
	if username != "admin" {
		t.Errorf("期望username=admin，实际=%s", username)
	}
}

func TestAdminUserUseCase_ValidateToken_Invalid(t *testing.T) {
	repo := &mockAdminUserRepo{
		validateTokenFn: func(ctx context.Context, token string) (bool, int64, string, error) {
			return false, 1, "", nil
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	valid, adminID, username, err := uc.ValidateToken(context.Background(), "invalid-token")

	if err != nil {
		t.Errorf("验证失败: %v", err)
	}
	if valid {
		t.Error("期望无效，实际有效")
	}
	if adminID != 1 {
		t.Errorf("期望adminID=1，实际=%d", adminID)
	}
	if username != "" {
		t.Errorf("期望username为空，实际=%s", username)
	}
}

func TestAdminUserUseCase_ValidateToken_Error(t *testing.T) {
	expectedErr := errors.New("Redis连接失败")
	repo := &mockAdminUserRepo{
		validateTokenFn: func(ctx context.Context, token string) (bool, int64, string, error) {
			return false, 0, "", expectedErr
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	_, _, _, err := uc.ValidateToken(context.Background(), "error-token")

	if err != expectedErr {
		t.Errorf("期望错误=%v，实际=%v", expectedErr, err)
	}
}

func TestAdminUserUseCase_Create_Success(t *testing.T) {
	repo := &mockAdminUserRepo{
		createFn: func(ctx context.Context, admin *AdminUser) (int64, error) {
			return 1, nil
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	id, err := uc.Create(context.Background(), &AdminUser{
		Username: "test",
		Password: "password",
		Name:     "测试用户",
	})

	if err != nil {
		t.Errorf("创建失败: %v", err)
	}
	if id != 1 {
		t.Errorf("期望ID=1，实际=%d", id)
	}
}

func TestAdminUserUseCase_Get_Success(t *testing.T) {
	repo := &mockAdminUserRepo{
		findByIDFn: func(ctx context.Context, id int64) (*AdminUser, error) {
			return &AdminUser{ID: id, Username: "admin"}, nil
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	admin, err := uc.Get(context.Background(), 1)

	if err != nil {
		t.Errorf("获取失败: %v", err)
	}
	if admin == nil {
		t.Fatal("管理员为空")
	}
	if admin.ID != 1 {
		t.Errorf("期望ID=1，实际=%d", admin.ID)
	}
}

func TestAdminUserUseCase_Get_NotFound(t *testing.T) {
	repo := &mockAdminUserRepo{
		findByIDFn: func(ctx context.Context, id int64) (*AdminUser, error) {
			return nil, nil
		},
	}

	uc := NewAdminUserUseCase(repo, "test-secret", log.NewStdLogger(nil))
	admin, err := uc.Get(context.Background(), 999)

	if err != nil {
		t.Errorf("获取失败: %v", err)
	}
	if admin != nil {
		t.Errorf("期望管理员为空，实际=%v", admin)
	}
}