package service

import (
	"context"
	"encoding/json"

	pb "cyi-app/api/service/user/v1"
	"cyi-app/app/service/user/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// MessageService 消息中心服务实现
type MessageService struct {
	pb.UnimplementedMessageServer
	uc  *biz.MessageUseCase
	log *log.Helper
}

// NewMessageService 创建 MessageService
func NewMessageService(uc *biz.MessageUseCase, logger log.Logger) *MessageService {
	return &MessageService{
		uc:  uc,
		log: log.NewHelper(logger),
	}
}

// CreateMessage 创建消息（管理后台发送通知）
func (s *MessageService) CreateMessage(ctx context.Context, req *pb.CreateMessageRequest) (*pb.CreateMessageReply, error) {
	id, err := s.uc.CreateMessage(ctx, &biz.Message{
		Title:          req.Title,
		Content:        req.Content,
		MsgType:        req.MsgType,
		TargetType:     req.TargetType,
		TargetUserIDs:  req.TargetUserIds,
		CreatedBy:      req.CreatedBy,
		ConversationID: req.ConversationId,
	})
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.CreateMessageReply{Id: id}, nil
}

// AdminListMessage 管理后台消息列表
func (s *MessageService) AdminListMessage(ctx context.Context, req *pb.AdminListMessageRequest) (*pb.AdminListMessageReply, error) {
	items, total, err := s.uc.AdminListMessage(ctx, req.Page, req.PageSize, req.MsgType, req.Status, req.Keyword)
	if err != nil {
		return nil, convertBizError(err)
	}
	list := make([]*pb.MessageInfo, 0, len(items))
	for _, m := range items {
		list = append(list, messageBizToProto(m))
	}
	return &pb.AdminListMessageReply{Items: list, Total: total}, nil
}

// DeleteMessage 下线/删除消息
func (s *MessageService) DeleteMessage(ctx context.Context, req *pb.DeleteMessageRequest) (*pb.DeleteMessageReply, error) {
	if err := s.uc.DeleteMessage(ctx, req.Id); err != nil {
		return &pb.DeleteMessageReply{Success: false}, convertBizError(err)
	}
	return &pb.DeleteMessageReply{Success: true}, nil
}

// ListMyMessage 我的消息列表
func (s *MessageService) ListMyMessage(ctx context.Context, req *pb.ListMyMessageRequest) (*pb.ListMyMessageReply, error) {
	items, total, err := s.uc.ListMyMessage(ctx, req.UserId, req.Page, req.PageSize, req.MsgType)
	if err != nil {
		return nil, convertBizError(err)
	}
	unread, err := s.uc.UnreadMessageCount(ctx, req.UserId)
	if err != nil {
		s.log.Errorf("UnreadMessageCount failed, userID=%d: %v", req.UserId, err)
		unread = 0
	}
	list := make([]*pb.MessageInfo, 0, len(items))
	for _, m := range items {
		list = append(list, messageBizToProto(m))
	}
	return &pb.ListMyMessageReply{Items: list, Total: total, UnreadCount: unread}, nil
}

// UnreadMessageCount 我的未读消息数
func (s *MessageService) UnreadMessageCount(ctx context.Context, req *pb.UnreadMessageCountRequest) (*pb.UnreadMessageCountReply, error) {
	count, err := s.uc.UnreadMessageCount(ctx, req.UserId)
	if err != nil {
		return nil, convertBizError(err)
	}
	return &pb.UnreadMessageCountReply{Count: count}, nil
}

// MarkMessageRead 标记单条消息已读
func (s *MessageService) MarkMessageRead(ctx context.Context, req *pb.MarkMessageReadRequest) (*pb.MarkMessageReadReply, error) {
	if err := s.uc.MarkMessageRead(ctx, req.UserId, req.MessageId); err != nil {
		return &pb.MarkMessageReadReply{Success: false}, convertBizError(err)
	}
	return &pb.MarkMessageReadReply{Success: true}, nil
}

// MarkAllMessageRead 全部标记已读
func (s *MessageService) MarkAllMessageRead(ctx context.Context, req *pb.MarkAllMessageReadRequest) (*pb.MarkAllMessageReadReply, error) {
	if err := s.uc.MarkAllMessageRead(ctx, req.UserId); err != nil {
		return &pb.MarkAllMessageReadReply{Success: false}, convertBizError(err)
	}
	return &pb.MarkAllMessageReadReply{Success: true}, nil
}

// messageBizToProto biz.Message -> pb.MessageInfo 转换
func messageBizToProto(m *biz.Message) *pb.MessageInfo {
	if m == nil {
		return nil
	}
	// 指定用户ID列表序列化为 JSON 字符串返回
	targetIDs := ""
	if len(m.TargetUserIDs) > 0 {
		if raw, err := json.Marshal(m.TargetUserIDs); err == nil {
			targetIDs = string(raw)
		}
	}
	return &pb.MessageInfo{
		Id:             m.ID,
		Title:          m.Title,
		Content:        m.Content,
		MsgType:        m.MsgType,
		IsRead:         m.IsRead,
		CreatedAt:      m.CreatedAt,
		CreatedBy:      m.CreatedBy,
		TargetType:     m.TargetType,
		TargetUserIds:  targetIDs,
		Status:         m.Status,
		ReadAt:         m.ReadAt,
		ConversationId: m.ConversationID,
	}
}
