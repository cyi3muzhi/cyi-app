package service

import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	service "cyi-app/pkg/kratos/service"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

// AiChat AI 聊天（非流式）
func (s *CinInterface) AiChat(ctx context.Context, req *v1.AiChatRequest) (*v1.AiChatReply, error) {
	message, err := s.bc.Chat(ctx, req.Message, req.SessionId)
	if err != nil {
		s.log.Errorf("ai chat failed: %v", err)
		return nil, err
	}
	return &v1.AiChatReply{
		Message: message,
	}, nil
}

// ChatSSE SSE 流式聊天（被 HandleFunc 注册到 /ai/v1/chat/stream）
func (s *CinInterface) ChatSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	userID, err := service.GetUserID(r.Context())
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	_ = userID // 已通过认证，具体 ID 不需要打印

	var msg wsMessage
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		writeSSE(w, "error", `{"error":"invalid request body"}`)
		return
	}
	if msg.Message == "" {
		writeSSE(w, "error", `{"error":"message is empty"}`)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeSSE(w, "error", `{"error":"streaming not supported"}`)
		return
	}

	_, err = s.bc.ChatStream(r.Context(), msg.Message, msg.SessionID, func(chunk string) {
		if chunk == "" {
			return
		}
		data, _ := json.Marshal(map[string]string{"chunk": chunk})
		writeSSE(w, "message", string(data))
		flusher.Flush()
	})
	if err != nil {
		s.log.Errorf("chat stream failed: %v", err)
		// 不向客户端回显内部错误细节，避免泄露内部信息
		writeSSE(w, "error", `{"error":"ai service unavailable"}`)
	}
	writeSSE(w, "done", `{"done":true}`)
	flusher.Flush()
}

func writeSSE(w http.ResponseWriter, event string, data string) {
	fmt.Fprintf(w, "event: %s\n", event)
	fmt.Fprintf(w, "data: %s\n\n", data)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应限制
	},
}

// wsMessage WebSocket 消息结构
type wsMessage struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

// wsChunk WebSocket 流式块
type wsChunk struct {
	Chunk string `json:"chunk,omitempty"`
	Done  bool   `json:"done,omitempty"`
	Error string `json:"error,omitempty"`
}

// ChatWS WebSocket 流式聊天
func (s *CinInterface) ChatWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Errorf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	for {
		// 读取用户消息
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg wsMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			conn.WriteJSON(wsChunk{Error: "invalid message format"})
			continue
		}
		if msg.Message == "" {
			continue
		}

		// 流式发送 AI 响应
		_, err = s.bc.ChatStream(r.Context(), msg.Message, msg.SessionID, func(chunk string) {
			if chunk == "" {
				return
			}
			conn.WriteJSON(wsChunk{Chunk: chunk})
		})
		if err != nil {
			s.log.Errorf("chat stream failed: %v", err)
			// 不向客户端回显内部错误细节，避免泄露内部信息
			conn.WriteJSON(wsChunk{Error: "ai service unavailable"})
		}
		conn.WriteJSON(wsChunk{Done: true})
	}
}
