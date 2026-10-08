package service

import (
	"context"
	"encoding/json"
	nethttp "net/http"

	"cyi-app/pkg/kratos/server/ws"

	"github.com/gorilla/websocket"
)

// wsSendData 客户端 send 帧数据
type wsSendData struct {
	ConversationID int64  `json:"conversation_id"`
	MsgType        int32  `json:"msg_type"`
	Content        string `json:"content"`
	Extra          string `json:"extra"`
}

// wsConvData 客户端 read/close 帧数据
type wsConvData struct {
	ConversationID int64 `json:"conversation_id"`
}

// wsErrorData 服务端 error 帧数据
type wsErrorData struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

// IMChatWS 用户端 IM WebSocket 聊天入口（/ws/chat/v2?token=xxx）
// 鉴权、升级连接、连接注册与读写循环由 ws.Serve 统一处理，业务帧在此处理。
// 连接注册/心跳续期/注销时维护 Redis 用户在线状态。
func (s *CinInterface) IMChatWS(w nethttp.ResponseWriter, r *nethttp.Request) {
	ws.Serve(w, r, []byte(s.auth.ApiKey), ws.KindUser,
		func(conn *websocket.Conn, id int64) *ws.Client {
			c := ws.NewUserClient(conn, id)
			key := s.chatUC.OnlineKey(id)
			c.OnlineRefresh = func() { s.chatUC.MarkUserOnline(key) }
			return c
		},
		func(c *ws.Client) {
			// 注册进 Hub，使 NATS 订阅的推送能匹配到该用户连接
			s.chatUC.Hub().Register(c)
			s.chatUC.MarkUserOnline(s.chatUC.OnlineKey(c.ID()))
		},
		func(c *ws.Client) {
			s.chatUC.Hub().Unregister(c)
			s.chatUC.MarkUserOffline(s.chatUC.OnlineKey(c.ID()))
		},
		s.onUserWSFrame,
	)
}

// onUserWSFrame 处理用户端一条业务帧
// 支持的帧：ping / create / send / read / close；返回 false 时断开连接
func (s *CinInterface) onUserWSFrame(c *ws.Client, userID int64, data []byte) bool {
	var frame ws.ClientFrame
	if err := json.Unmarshal(data, &frame); err != nil {
		c.Send(ws.NewServerFrame("error", wsErrorData{Code: 1000, Message: "无效的消息格式"}))
		return true
	}

	ctx := context.Background()
	switch frame.Type {
	case "ping":
		// 心跳由 WritePump 负责，客户端 ping 直接忽略
		return true
	case "create":
		conv, err := s.chatUC.CreateConversation(ctx, userID)
		if err != nil {
			s.writeWSError(c, err)
			return true
		}
		c.Send(ws.NewServerFrame("conversation_created", map[string]interface{}{"conversation": conv}))
		return true
	case "send":
		var d wsSendData
		if err := json.Unmarshal(frame.Data, &d); err != nil || d.ConversationID <= 0 {
			c.Send(ws.NewServerFrame("error", wsErrorData{Code: 1001, Message: "send 帧参数无效"}))
			return true
		}
		reply, err := s.chatUC.SendMessage(ctx, userID, d.ConversationID, d.MsgType, d.Content, d.Extra)
		if err != nil {
			s.writeWSError(c, err)
			return true
		}
		c.Send(ws.NewServerFrame("send_ack", map[string]interface{}{
			"conversation_id": d.ConversationID,
			"message_id":      reply.MessageId,
			"created_at":      reply.CreatedAt,
		}))
		return true
	case "read":
		var d wsConvData
		if err := json.Unmarshal(frame.Data, &d); err != nil || d.ConversationID <= 0 {
			c.Send(ws.NewServerFrame("error", wsErrorData{Code: 1001, Message: "read 帧参数无效"}))
			return true
		}
		if err := s.chatUC.MarkConversationRead(ctx, userID, d.ConversationID); err != nil {
			s.writeWSError(c, err)
			return true
		}
		c.Send(ws.NewServerFrame("read_ack", map[string]interface{}{"conversation_id": d.ConversationID}))
		return true
	case "close":
		var d wsConvData
		if err := json.Unmarshal(frame.Data, &d); err != nil || d.ConversationID <= 0 {
			c.Send(ws.NewServerFrame("error", wsErrorData{Code: 1001, Message: "close 帧参数无效"}))
			return true
		}
		if err := s.chatUC.CloseConversation(ctx, userID, d.ConversationID); err != nil {
			s.writeWSError(c, err)
			return true
		}
		c.Send(ws.NewServerFrame("close_ack", map[string]interface{}{"conversation_id": d.ConversationID}))
		return true
	default:
		c.Send(ws.NewServerFrame("error", wsErrorData{Code: 1000, Message: "未知的消息类型"}))
		return true
	}
}

// writeWSError 将 gRPC 错误转换为 error 帧推送给客户端（不泄露内部细节）
func (s *CinInterface) writeWSError(c *ws.Client, err error) {
	if err == nil {
		return
	}
	s.log.Warnf("ws chat operation failed: %v", err)
	c.Send(ws.NewServerFrame("error", wsErrorData{Code: 1100, Message: "操作失败，请稍后重试"}))
}
