package service

import (
	"context"

	pb "cyi-app/api/blades/v1"
	"cyi-app/app/blades/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

type BladesService struct {
	pb.UnimplementedBladesServer
	uc  *biz.BladesUseCase
	log *log.Helper
}

func NewBladesService(uc *biz.BladesUseCase, logger log.Logger) *BladesService {
	return &BladesService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

func (s *BladesService) Chat(ctx context.Context, req *pb.ChatRequest) (*pb.ChatReply, error) {
	response, err := s.uc.Chat(ctx, req.Message)
	if err != nil {
		s.log.Error("Chat failed: %v", err)
		return nil, err
	}
	return &pb.ChatReply{
		Message: response,
	}, nil
}
