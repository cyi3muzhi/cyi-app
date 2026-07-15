package biz

import (
	"context"
	"strings"
	"time"

	v1 "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/pkg/encryption"

	"github.com/go-kratos/kratos/v2/log"
)

// AdminUser 管理员用户
type AdminUser struct {
	ID         int64
	Username   string
	Password   string
	Name       string
	Role       string
	Email      string
	Phone      string
	Department string
	Position   string
	Bio        string
	Avatar     string
	Status     int32 // 1=正常 2=冻结
	LastLogin  int64
	CreatedAt  time.Time
}

// AdminUserRepo 管理员数据仓库接口
type AdminUserRepo interface {
	FindByUsername(ctx context.Context, username string) (*AdminUser, error)
	FindByID(ctx context.Context, id int64) (*AdminUser, error)
	Create(ctx context.Context, admin *AdminUser) (int64, error)
	Update(ctx context.Context, admin *AdminUser) error
	List(ctx context.Context, page, pageSize int32, keyword, status, role string) ([]*AdminUser, int64, error)
	ToggleStatus(ctx context.Context, id int64) (int32, error)
	UpdateLastLogin(ctx context.Context, id int64) error
	VerifyPassword(ctx context.Context, id int64, password string) error
	SaveToken(ctx context.Context, token string, adminID int64) error
	RemoveToken(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (bool, int64, string, error)
	UpdateAvatar(ctx context.Context, adminID int64, avatarURL string) error
}

// AdminUserUseCase 管理员业务逻辑
type AdminUserUseCase struct {
	repo   AdminUserRepo
	apiKey string
	log    *log.Helper
}

// NewAdminUserUseCase 创建 AdminUserUseCase
func NewAdminUserUseCase(repo AdminUserRepo, apiKey string, logger log.Logger) *AdminUserUseCase {
	return &AdminUserUseCase{
		repo:   repo,
		apiKey: apiKey,
		log:    log.NewHelper(logger),
	}
}

// Login 管理员登录验证，验证通过返回 admin 信息（token 由 admin BFF 生成）
func (uc *AdminUserUseCase) Login(ctx context.Context, username, password string) (*AdminUser, error) {
	admin, err := uc.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, v1.ErrorUserNotFound("管理员账号不存在")
	}
	if admin == nil {
		return nil, v1.ErrorUserNotFound("管理员账号不存在")
	}
	if admin.Status == 2 {
		return nil, v1.ErrorUserNotFound("管理员账号已冻结")
	}
	if err := uc.repo.VerifyPassword(ctx, admin.ID, password); err != nil {
		return nil, v1.ErrorPasswordError("管理员密码错误")
	}
	_ = uc.repo.UpdateLastLogin(ctx, admin.ID)
	return admin, nil
}

// Create 创建管理员
func (uc *AdminUserUseCase) Create(ctx context.Context, admin *AdminUser) (int64, error) {
	encryptedPassword, err := encryption.EncryptPassword(admin.Password)
	if err != nil {
		return 0, err
	}
	admin.Password = encryptedPassword
	id, err := uc.repo.Create(ctx, admin)
	if err != nil {
		// 用户名重复 → 转换为业务错误
		if strings.Contains(err.Error(), "Duplicate entry") && strings.Contains(err.Error(), "idx_admin_user_username") {
			return 0, v1.ErrorUserAlreadyRegistered("该账号已注册，请更换用户名")
		}
		return 0, err
	}
	return id, nil
}

// Update 更新管理员
func (uc *AdminUserUseCase) Update(ctx context.Context, admin *AdminUser) error {
	return uc.repo.Update(ctx, admin)
}

// Get 获取管理员
func (uc *AdminUserUseCase) Get(ctx context.Context, id int64) (*AdminUser, error) {
	return uc.repo.FindByID(ctx, id)
}

// List 管理员列表
func (uc *AdminUserUseCase) List(ctx context.Context, page, pageSize int32, keyword, status, role string) ([]*AdminUser, int64, error) {
	return uc.repo.List(ctx, page, pageSize, keyword, status, role)
}

// ToggleStatus 切换状态
func (uc *AdminUserUseCase) ToggleStatus(ctx context.Context, id int64) (int32, error) {
	return uc.repo.ToggleStatus(ctx, id)
}

// UpdateAvatar 更新管理员头像
func (uc *AdminUserUseCase) UpdateAvatar(ctx context.Context, adminID int64, avatarURL string) error {
	return uc.repo.UpdateAvatar(ctx, adminID, avatarURL)
}

// 确保 import 不被移除
var _ = &v1.AdminLoginReply{}
