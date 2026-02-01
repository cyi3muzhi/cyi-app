package biz

// 领域层
import (
	"context"
	"cyi-app/app/cin/interface/internal/conf"

	v1 "cyi-app/api/cin/interface/v1"
	usv1 "cyi-app/api/service/user/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	server "cyi-app/pkg/kratos/server/middleware"
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

func (uc *AuthUseCase) Register(ctx context.Context, in *User) error {

	// 业务逻辑
	// 查看用户是否存在
	user, err := uc.userRepo.FindByUsername(ctx, in.Username)
	if err == nil && user != nil && user.ID != 0 {
		uc.log.Infof("User already exists: %v", user)
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
	if err := uc.userRepo.Save(ctx, in); err != nil {
		uc.log.Errorf("Failed to save user: %v", err)
		return v1.ErrorUserRegisterFailed("save user failed")
	}
	return nil
}

func (uc *AuthUseCase) Login(ctx context.Context, in *User) (*v1.LoginReply, error) {

	// 业务逻辑
	// 查看用户是否存在
	user, err := uc.userRepo.FindByUsername(ctx, in.Username)
	if err != nil {
		// 用户不存在
		uc.log.Errorf("User not found: %v", err)
		return nil, v1.ErrorUserNotFound("user not found")
	}
	// check permission(password blacklist etc...)
	err = uc.userRepo.VerifyPassword(ctx, user, in.Password)
	if err != nil {
		// 密码错误
		uc.log.Errorf("Password error: %v", err)
		return nil, v1.ErrorPasswordError("password not match")
	}
	// generate token
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, server.UserClaims{
		ID: user.ID,
		Username: user.Username,
	})
	signedString, err := claims.SignedString([]byte(uc.key))
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		Token: signedString,
	}, nil
}
