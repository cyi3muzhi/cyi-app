package biz

// 领域层
import (
	"context"
	"cyi-app/app/cin/interface/internal/conf"

	v1 "cyi-app/api/cin/interface/v1"
	usv1 "cyi-app/api/service/user/v1"

	"github.com/go-kratos/kratos/v2/log"
)

type AuthUseCase struct {
	log      *log.Helper
	key      string
	userRepo UserRepo
}

func NewAuthUseCase(userRepo UserRepo, authConf *conf.Auth, logger log.Logger) *AuthUseCase {
	return &AuthUseCase{
		log:      log.NewHelper(logger),
		key:      authConf.ApiKey,
		userRepo: userRepo,
	}
}

// 领域对象
type User struct {
	ID int64 `json:"id"`
	// 密码
	Password string `json:"password"`
	// 用户名
	Username string `json:"username"`
}

func (uc *AuthUseCase) Register(ctx context.Context, user *User) error {

	// 业务逻辑
	// 查看用户是否存在
	existingUser, err := uc.userRepo.FindByUsername(ctx, user.Username)
	if err == nil && existingUser != nil && existingUser.ID != 0 {
		uc.log.Infof("User already exists: %v", existingUser)
		// 用户已存在
		return v1.ErrorUserAlreadyRegistered("username already exists")
	}

	// 检查是否是其他错误（非用户不存在错误）
	if err != nil && !usv1.IsUserNotFound(err) {
		// 记录错误日志
		uc.log.Errorf("Failed to check user existence: %v", err)
		return v1.ErrorUserRegisterFailed("check user existence failed")
	}

	// save user
	if err := uc.userRepo.Save(ctx, user); err != nil {
		uc.log.Errorf("Failed to save user: %v", err)
		return v1.ErrorUserRegisterFailed("save user failed")
	}
	return nil
}

func (u *AuthUseCase) Login(ctx context.Context, user *User) (*User, error) {

	return nil, nil
}
