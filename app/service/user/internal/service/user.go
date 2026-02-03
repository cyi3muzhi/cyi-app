package service

import (
	"context"
	"fmt"

	pb "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/biz"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
)

type UserService struct {
	pb.UnimplementedUserServer
	uc  *biz.UserUseCase
	log *log.Helper
}

func NewUserService(uc *biz.UserUseCase, logger log.Logger) *UserService {
	return &UserService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserReply, error) {
	u := &biz.User{}
	if err := convert.Convert(req, u); err != nil {
		return nil, err
	}
	user, err := s.uc.CreateUser(ctx, u)
	if err != nil {
		return nil, err
	}
	reply := &pb.CreateUserReply{}
	if err := convert.Convert(user, reply); err != nil {
		return nil, err
	}

	return reply, nil
}
func (s *UserService) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserReply, error) {
	return &pb.UpdateUserReply{}, nil
}
func (s *UserService) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*pb.DeleteUserReply, error) {
	return &pb.DeleteUserReply{}, nil
}
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserReply, error) {
	user, err := s.uc.GetUser(ctx, &biz.User{
		ID: req.Id,
	})
	if err != nil {
		return nil, err
	}
	reply := &pb.GetUserReply{}
	if err := convert.Convert(user, reply); err != nil {
		return nil, err
	}
	return reply, nil
}
func (s *UserService) ListUser(ctx context.Context, req *pb.ListUserRequest) (*pb.ListUserReply, error) {
	return &pb.ListUserReply{}, nil
}

// 校验用户是否存在
func (s *UserService) CheckUserExist(ctx context.Context, req *pb.CheckUserExistRequest) (*pb.CheckUserExistReply, error) {
	fmt.Println("phone:", req.Phone)
	// 查询用户是否存在
	user, err := s.uc.GetUser(ctx, &biz.User{
		Phone: req.Phone,
	})
	if err != nil {
		return nil, err
	}

	fmt.Println("1")
	if user.ID != 0 {
		return &pb.CheckUserExistReply{
			Exist: true,
		}, nil
	}
	// 查询用户是否存在
	user, err = s.uc.GetUser(ctx, &biz.User{
		Username: req.Username,
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("2")

	if user.ID != 0 {
		return &pb.CheckUserExistReply{
			Exist: true,
		}, nil
	}
	fmt.Println("3")

	return &pb.CheckUserExistReply{
		Exist: false,
	}, nil
}

func (s *UserService) GetUserByUsername(ctx context.Context, in *pb.GetUserByUsernameReq) (*pb.GetUserByUsernameReply, error) {
	user, err := s.uc.GetUserByUsername(ctx, in)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserByUsernameReply{
		Id:       user.ID,
		Username: user.Username,
	}, nil
}

// 校验密码
func (s *UserService) VerifyPassword(ctx context.Context, in *pb.VerifyPasswordRequest) (*pb.VerifyPasswordReply, error) {
	u := &biz.User{}
	if err := convert.Convert(in, u); err != nil {
		return nil, err
	}

	if err := s.uc.VerifyPassword(ctx, u); err != nil {
		return &pb.VerifyPasswordReply{Valid: false}, err
	}
	return &pb.VerifyPasswordReply{Valid: true}, nil
}
