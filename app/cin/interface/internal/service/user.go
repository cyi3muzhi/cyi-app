package service

import (
	"context"

	pb "cyi-app/api/cin/interface/user"
	"cyi-app/app/cin/interface/internal/biz"

	"github.com/go-kratos/kratos/v2/log"

	kerr "cyi-app/api/err"
)

type UserService struct {
	pb.UnimplementedUserServer
	log *log.Helper
	uc  *biz.UserUseCase
}

func NewUserService(uc *biz.UserUseCase, logger log.Logger) *UserService {
	return &UserService{
		log: log.NewHelper(logger),
		uc:  uc,
	}
}

func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	return &pb.LoginReply{Token: "token"}, nil
}
func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	// 校验密码是否输入一致
	if req.Password != req.ConfirmPassword {
		return nil, kerr.ErrorPasswordNotMatch("密码不一致")
	}
	u := &biz.User{
		Phone:    req.Phone,
		Password: req.Password,
		Username: req.Username,
	}
	ok, err := s.uc.Register(ctx, u)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterReply{
		Ok: ok,
	}, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoReply, error) {
	user, err := s.uc.GetUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserInfoReply{
		Id:   user.ID, // 使用正确的字段名ID
		Name: user.Name,
	}, nil
}
