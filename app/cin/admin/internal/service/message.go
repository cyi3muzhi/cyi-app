package service

import (
	"context"
	"encoding/json"

	adminv1 "cyi-app/api/cin/admin/v1"
	"cyi-app/app/cin/admin/internal/biz"

	kratosService "cyi-app/pkg/kratos/service"
)

// SendMessage 发送消息（全体/指定用户）
func (s *AdminService) SendMessage(ctx context.Context, req *adminv1.SendMessageRequest) (*adminv1.SendMessageReply, error) {
	// 获取当前管理员ID作为创建人
	adminID, err := kratosService.GetUserID(ctx)
	if err != nil {
		adminID = 0
	}
	s.log.Infof("发送消息: title=%s, msgType=%d, targetType=%d", req.Title, req.MsgType, req.TargetType)

	id, err := s.msgUC.SendMessage(ctx, &biz.Message{
		Title:         req.Title,
		Content:       req.Content,
		MsgType:       req.MsgType,
		TargetType:    req.TargetType,
		TargetUserIDs: req.TargetUserIds,
		CreatedBy:     adminID,
	})
	if err != nil {
		s.log.Errorf("发送消息失败: %v", err)
		return nil, err
	}
	return &adminv1.SendMessageReply{Id: id}, nil
}

// ListMessage 消息列表
func (s *AdminService) ListMessage(ctx context.Context, req *adminv1.ListMessageRequest) (*adminv1.ListMessageReply, error) {
	reply, err := s.msgUC.ListMessage(ctx, &biz.MessageListRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
		MsgType:  req.MsgType,
		Status:   req.Status,
		Keyword:  req.Keyword,
	})
	if err != nil {
		s.log.Errorf("查询消息列表失败: %v", err)
		return nil, err
	}

	items := make([]*adminv1.MessageItem, 0, len(reply.Items))
	for _, m := range reply.Items {
		items = append(items, messageBizToProto(m))
	}
	return &adminv1.ListMessageReply{
		Items: items,
		Total: reply.Total,
	}, nil
}

// DeleteMessage 下线/删除消息
func (s *AdminService) DeleteMessage(ctx context.Context, req *adminv1.DeleteMessageRequest) (*adminv1.DeleteMessageReply, error) {
	s.log.Infof("下线消息: id=%d", req.Id)
	if err := s.msgUC.DeleteMessage(ctx, req.Id); err != nil {
		s.log.Errorf("下线消息失败: %v", err)
		return &adminv1.DeleteMessageReply{Success: false}, err
	}
	return &adminv1.DeleteMessageReply{Success: true}, nil
}

// messageBizToProto biz.Message -> adminv1.MessageItem 转换
func messageBizToProto(m *biz.Message) *adminv1.MessageItem {
	if m == nil {
		return nil
	}
	return &adminv1.MessageItem{
		Id:            m.ID,
		Title:         m.Title,
		Content:       m.Content,
		MsgType:       m.MsgType,
		CreatedAt:     m.CreatedAt,
		CreatedBy:     m.CreatedBy,
		TargetType:    m.TargetType,
		TargetUserIds: int64SliceToJSON(m.TargetUserIDs),
		Status:        m.Status,
	}
}

// int64SliceToJSON 将 int64 切片序列化为 JSON 字符串（用于 proto string 字段）
func int64SliceToJSON(ids []int64) string {
	if ids == nil {
		return "[]"
	}
	raw, err := json.Marshal(ids)
	if err != nil {
		return "[]"
	}
	return string(raw)
}
