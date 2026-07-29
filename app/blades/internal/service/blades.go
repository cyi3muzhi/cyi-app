package service

import (
	"context"
	"io"

	pb "cyi-app/api/blades/v1"
	"cyi-app/app/blades/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// BladesService 实现 Blades 服务接口
type BladesService struct {
	pb.UnimplementedBladesServer
	uc     *biz.BladesUseCase
	agent  *biz.AgentUseCase
	log    *log.Helper
}

// NewBladesService 创建 BladesService
func NewBladesService(uc *biz.BladesUseCase, agent *biz.AgentUseCase, logger log.Logger) *BladesService {
	return &BladesService{
		uc:    uc,
		agent: agent,
		log:   log.NewHelper(logger),
	}
}

// Chat 非流式聊天
func (s *BladesService) Chat(ctx context.Context, req *pb.ChatRequest) (*pb.ChatReply, error) {
	if s.agent == nil {
		// 降级使用原有的简单聊天
		response, err := s.uc.Chat(ctx, req.Message)
		if err != nil {
			s.log.Error("Chat failed: %v", err)
			return nil, err
		}
		return &pb.ChatReply{Message: response}, nil
	}

	response, err := s.agent.Chat(ctx, req.AdminId, req.Message)
	if err != nil {
		s.log.Error("Agent Chat failed: %v", err)
		return nil, err
	}
	return &pb.ChatReply{Message: response}, nil
}

// ChatStream 流式聊天（服务端流式）
func (s *BladesService) ChatStream(req *pb.ChatStreamRequest, stream pb.Blades_ChatStreamServer) error {
	if s.agent == nil {
		return nil
	}

	err := s.agent.ChatStream(stream.Context(), req.AdminId, req.Message, func(chunk string, done bool) {
		resp := &pb.ChatChunkReply{
			Chunk: chunk,
			Done:  done,
		}
		if err := stream.Send(resp); err != nil && err != io.EOF {
			s.log.Error("ChatStream Send failed: %v", err)
		}
	})
	if err != nil {
		s.log.Error("Agent ChatStream failed: %v", err)
		return err
	}
	return nil
}

// GetChatHistory 获取聊天历史
func (s *BladesService) GetChatHistory(ctx context.Context, req *pb.GetChatHistoryRequest) (*pb.GetChatHistoryReply, error) {
	if s.agent == nil {
		return &pb.GetChatHistoryReply{Messages: []*pb.HistoryMessage{}}, nil
	}

	history, err := s.agent.LoadChatHistory(ctx, req.AdminId)
	if err != nil {
		s.log.Error("GetChatHistory failed: %v", err)
		return nil, err
	}

	messages := make([]*pb.HistoryMessage, 0, len(history))
	for _, m := range history {
		messages = append(messages, &pb.HistoryMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		})
	}
	return &pb.GetChatHistoryReply{Messages: messages}, nil
}

// ClearChatHistory 清除聊天历史
func (s *BladesService) ClearChatHistory(ctx context.Context, req *pb.ClearChatHistoryRequest) (*pb.ClearChatHistoryReply, error) {
	if s.agent == nil {
		return &pb.ClearChatHistoryReply{Success: true}, nil
	}

	err := s.agent.ClearHistory(ctx, req.AdminId)
	if err != nil {
		s.log.Error("ClearChatHistory failed: %v", err)
		return nil, err
	}
	return &pb.ClearChatHistoryReply{Success: true}, nil
}
