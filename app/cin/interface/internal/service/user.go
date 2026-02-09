package service

// 应用层
// 业务编排 DTO
import (
	"context"
	v1 "cyi-app/api/cin/interface/v1"
	"cyi-app/app/cin/interface/internal/biz"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cyi-app/pkg/kratos/server/middleware"
	service "cyi-app/pkg/kratos/service"
	"cyi-app/pkg/util/convert"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *CinInterface) Register(ctx context.Context, req *v1.RegisterReq) (*v1.RegisterReply, error) {
	in := &biz.User{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	err := s.ac.Register(ctx, in)
	if err != nil {
		return nil, err
	}

	return &v1.RegisterReply{
		Success: true,
	}, nil
}

func (s *CinInterface) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	in := &biz.User{}
	if err := convert.Convert(req, in); err != nil {
		return nil, err
	}
	return s.ac.Login(ctx, in)
}

// 获取用户信息
func (s *CinInterface) UserInfo(ctx context.Context, _ *emptypb.Empty) (*v1.UserInfoReply, error) {
	// 获取上下文的用户id
	id, err := service.GetUserID(ctx)
	if err != nil {
		return nil, err
	}

	// 获取用户信息
	user, err := s.uc.UserInfo(ctx, id)
	if err != nil {
		return nil, err
	}
	reply := &v1.UserInfoReply{}
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源的连接（生产环境应限制）
	},
}

// Message 定义消息结构
type Message struct {
	Type       string `json:"type"`        // 消息类型："private" 或 "group"
	SenderID   int64  `json:"sender_id"`   // 发送者 ID
	ReceiverID int64  `json:"receiver_id"` // 接收者 ID（私聊）
	GroupID    int64  `json:"group_id"`    // 群组 ID（群聊）
	Content    string `json:"content"`     // 消息内容
	Timestamp  int64  `json:"timestamp"`   // 时间戳
}

// registerUserConnection 注册用户连接
func (s *CinInterface) registerUserConnection(userID int64, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userConnections[userID] = conn
	s.log.Infof("用户 %d 连接已注册", userID)
}

// unregisterUserConnection 注销用户连接
func (s *CinInterface) unregisterUserConnection(userID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.userConnections, userID)
	s.log.Infof("用户 %d 连接已注销", userID)
}

// handleMessage 处理消息
func (s *CinInterface) handleMessage(msg *Message) {
	switch msg.Type {
	case "private":
		s.sendPrivateMessage(msg)
	case "group":
		s.sendGroupMessage(msg)
	default:
		s.log.Errorf("未知消息类型: %s", msg.Type)
	}
}

// sendPrivateMessage 发送私聊消息
func (s *CinInterface) sendPrivateMessage(msg *Message) {
	s.mu.RLock()
	conn, ok := s.userConnections[msg.ReceiverID]
	s.mu.RUnlock()

	if !ok {
		s.log.Errorf("用户 %d 未连接", msg.ReceiverID)
		return
	}

	response, err := json.Marshal(msg)
	if err != nil {
		s.log.Errorf("序列化消息失败: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
		s.log.Errorf("发送私聊消息失败: %v", err)
	}
}

// sendGroupMessage 发送群聊消息
func (s *CinInterface) sendGroupMessage(msg *Message) {
	s.mu.RLock()
	members, ok := s.groupMembers[msg.GroupID]
	s.mu.RUnlock()

	if !ok {
		s.log.Errorf("群组 %d 不存在", msg.GroupID)
		return
	}

	response, err := json.Marshal(msg)
	if err != nil {
		s.log.Errorf("序列化消息失败: %v", err)
		return
	}

	for _, memberID := range members {
		// 跳过发送者自己
		if memberID == msg.SenderID {
			continue
		}

		s.mu.RLock()
		conn, ok := s.userConnections[memberID]
		s.mu.RUnlock()

		if ok {
			if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
				s.log.Errorf("发送群聊消息失败: %v", err)
			}
		}
	}
}

// parseTokenFromRequest 从请求中解析 token 并返回用户 ID
func (s *CinInterface) parseTokenFromRequest(r *http.Request) (int64, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return 0, fmt.Errorf("token 为空")
	}

	// 移除 Bearer 前缀
	tokenStr := token
	if len(token) > 7 && token[:7] == "Bearer " {
		tokenStr = token[7:]
	}

	// 检查 token 格式
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("token 格式错误，应该包含3个部分，实际包含: %d", len(parts))
	}

	// 解析 token
	claims := &middleware.UserClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// 使用与中间件相同的密钥
		// 注意：这里应该从配置中获取，而不是硬编码
		return []byte(s.auth.ApiKey), nil
	})
	if err != nil {
		return 0, fmt.Errorf("token 解析失败: %v", err)
	}

	// 返回用户 ID
	return claims.ID, nil
}

// 处理聊天
func (s *CinInterface) UserChat(w http.ResponseWriter, r *http.Request) {

	// 解析 token 获取用户 ID
	userID, err := s.parseTokenFromRequest(r)
	if err != nil {
		s.log.Errorf("解析 token 失败: %v", err)
		return
	}
	s.log.Infof("用户id: %d", userID)

	// 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.Errorf("WebSocket 升级失败: %v", err)
		return
	}
	s.log.Infof("WebSocket 连接成功")
	defer conn.Close()

	// 注册用户连接
	s.registerUserConnection(userID, conn)
	defer s.unregisterUserConnection(userID)

	// 处理 WebSocket 消息
	s.log.Infof("开始处理 WebSocket 消息")
	for {
		// 读取消息
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			s.log.Errorf("WebSocket 读取失败: %v", err)
			break
		}
		s.log.Infof("messageType: %d", messageType)
		if messageType == websocket.TextMessage {
			// 处理文本消息
			s.log.Infof("收到文本消息: %s", p)

			// 解析消息结构
			var msg Message
			if err := json.Unmarshal(p, &msg); err != nil {
				s.log.Errorf("解析消息失败: %v", err)
				continue
			}

			// 设置发送者 ID
			msg.SenderID = userID
			msg.Timestamp = time.Now().Unix()

			// 根据消息类型处理
			s.handleMessage(&msg)

			// 回显消息给发送者
			response, _ := json.Marshal(msg)
			if err := conn.WriteMessage(messageType, response); err != nil {
				s.log.Errorf("WebSocket 写入失败: %v", err)
				break
			}
		}
	}
}
