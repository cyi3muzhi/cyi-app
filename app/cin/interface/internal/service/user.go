package service

// 应用层
// 业务编排 DTO
import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
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
