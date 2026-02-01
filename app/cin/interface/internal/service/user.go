package service

// 应用层
// 业务编排 DTO
import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"

	service "cyi-app/pkg/kratos/service"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *CinInterface) Register(ctx context.Context, req *v1.RegisterReq) (*v1.RegisterReply, error) {
	err := s.ac.Register(ctx, &biz.User{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &v1.RegisterReply{
		Success: true,
	}, nil
}

func (s *CinInterface) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	reply, err := s.ac.Login(ctx, &biz.User{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LoginReply{
		Token: reply.Token,
	}, nil
}

// 获取用户信息
func (s *CinInterface) UserInfo(ctx context.Context, _ *emptypb.Empty) (*v1.UserInfoReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	s.log.Infof("UserInfo: %v", id)

	// 获取用户信息
	user, err := s.uc.UserInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	return &v1.UserInfoReply{
		Id:       user.ID,
		Username: user.Username,
	}, nil
}
