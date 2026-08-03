package biz

// 领域层
import (
	"context"
	"cyi-app/app/cin/interface/internal/conf"
	"time"

	v1 "cyi-app/api/cin/interface/v1"

	server "cyi-app/pkg/kratos/server/middleware"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	// 用户名（登录账号）
	Username string `json:"username"`
	// 展示名
	Nickname string `json:"nickname"`
	// 手机号
	Phone string `json:"phone"`
	// 头像URL
	Avatar string `json:"avatar"`
	// 性别 0=未知 1=男 2=女
	Gender int32 `json:"gender"`
	// 密码
	Password string `json:"password"`
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
		ID:       user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			Issuer:    "cin",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	})
	signedString, err := claims.SignedString([]byte(uc.key))
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		Token: signedString,
	}, nil
}
