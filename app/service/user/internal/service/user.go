package service

import (
	"context"
	"errors"
	"fmt"

	pb "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/biz"
	"cyi-app/pkg/util/convert"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
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
	u := &biz.User{}
	if err := convert.Convert(req, u); err != nil {
		return nil, err
	}
	if err := s.uc.UpdateUser(ctx, u); err != nil {
		return &pb.UpdateUserReply{
			Success: false,
		}, err
	}
	return &pb.UpdateUserReply{
		Success: true,
	}, nil
}
// DeleteUser 删除用户
// TODO: 待业务需求明确后实现
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
// ListUser 获取用户列表
// TODO: 待业务需求明确后实现
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
	reply := &pb.GetUserByUsernameReply{}
	if err := convert.Convert(user, reply); err != nil {
		return nil, err
	}
	return reply, nil
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

// 统计用户数量
func (s *UserService) UserCount(ctx context.Context, _ *emptypb.Empty) (*pb.UserCountReply, error) {
	count, err := s.uc.UserCount(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.UserCountReply{Count: count}, nil
}

// 注册用户
func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterReply, error) {
	u := &biz.User{}
	if err := convert.Convert(req, u); err != nil {
		return nil, err
	}
	if err := s.uc.Register(ctx, u); err != nil {
		return nil, convertBizError(err)
	}
	return &pb.RegisterReply{
		Success: true,
	}, nil
}

// convertBizError 将 biz 层错误转换为 proto 错误
func convertBizError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, biz.ErrUserNotFound) {
		return pb.ErrorUserNotFound("%s", err.Error())
	}
	if errors.Is(err, biz.ErrPasswordError) {
		return pb.ErrorPasswordError("%s", err.Error())
	}
	if errors.Is(err, biz.ErrUserAlreadyRegistered) {
		return pb.ErrorUserAlreadyRegistered("%s", err.Error())
	}
	if errors.Is(err, biz.ErrUserRegisterFailed) {
		return pb.ErrorUserRegisterFailed("%s", err.Error())
	}
	return err
}

func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginReply, error) {
	u := &biz.User{}
	if err := convert.Convert(req, u); err != nil {
		return nil, err
	}
	reply, err := s.uc.Login(ctx, u, req.ApiKey)
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.LoginReply{
		Token: reply.Token,
	}, nil
}
