package service

import (
	"context"

	pb "github.com/cin/cin-app/api/cin/interface/user"
	"github.com/cin/cin-app/app/cin/interface/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
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
	_, err := s.uc.GetUser(ctx, 1)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterReply{}, nil
}
