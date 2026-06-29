package service

// 应用层
// 业务编排 DTO
import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"

	service "cyi-app/pkg/kratos/service"
	"cyi-app/pkg/util/convert"

	"google.golang.org/protobuf/types/known/emptypb"
	com "cyi-app/api/common"
)

func (s *CinInterface) Register(ctx context.Context, req *v1.RegisterReq) (*v1.RegisterReply, error) {
	err := s.uc.Register(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.RegisterReply{
		Success: true,
	}, nil
}

func (s *CinInterface) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	return s.uc.Login(ctx, req)
}

// 获取用户信息
func (s *CinInterface) GetMyInfo(ctx context.Context, _ *emptypb.Empty) (*v1.GetMyInfoReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, com.ErrorUnauthorized("用户未登录")
	}

	// 获取用户信息
	user, err := s.uc.UserInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	reply := &v1.GetMyInfoReply{}
	if err := convert.Convert(user, reply); err != nil {
		return nil, err
	}
	return reply, nil
}

// 更新用户信息
func (s *CinInterface) UpdateUserInfo(ctx context.Context, req *v1.UpdateUserInfoRequest) (*v1.UpdateUserInfoReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}
	in := &biz.User{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	in.ID = id

	// 更新用户信息
	err = s.uc.UpdateUserInfo(ctx, in)
	if err != nil {
		return &v1.UpdateUserInfoReply{
			Success: false,
		}, err
	}
	reply := &v1.UpdateUserInfoReply{
		Success: true,
	}
	return reply, nil
}