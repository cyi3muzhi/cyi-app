package ws

import (
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait 写超时
	writeWait = 10 * time.Second
	// pongWait 读取心跳超时（客户端需在此前发 ping/pong）
	pongWait = 60 * time.Second
	// pingPeriod ping 发送周期（需小于 pongWait）
	pingPeriod = 30 * time.Second
	// sendBuffer 发送缓冲队列长度
	sendBuffer = 64
)

// Client 单个 WebSocket 连接
// 用户连接绑定 userID，客服连接绑定 agentID；
// 通过 channel 缓冲写入，推送不阻塞 Hub 分发。
type Client struct {
	conn *websocket.Conn
	send chan []byte
	kind Kind
	// 用户端角色：用户ID；客服端角色：客服ID（二者其一非零）
	userID  int64
	agentID int64
	// OnlineRefresh 心跳续期回调（可选）：每次发送 ping 时调用，
	// 网关可在此刷新 Redis 在线状态 TTL
	OnlineRefresh func()
}

// NewUserClient 创建用户连接
func NewUserClient(conn *websocket.Conn, userID int64) *Client {
	return &Client{conn: conn, send: make(chan []byte, sendBuffer), kind: KindUser, userID: userID}
}

// NewAgentClient 创建客服连接
func NewAgentClient(conn *websocket.Conn, agentID int64) *Client {
	return &Client{conn: conn, send: make(chan []byte, sendBuffer), kind: KindAgent, agentID: agentID}
}

// Kind 返回连接角色
func (c *Client) Kind() Kind { return c.kind }

// ID 返回连接绑定的主体 ID（用户端=用户ID，客服端=客服ID）
func (c *Client) ID() int64 {
	if c.kind == KindAgent {
		return c.agentID
	}
	return c.userID
}

// Send 非阻塞推入发送队列；队列满则丢弃（尽力而为）
func (c *Client) Send(data []byte) bool {
	select {
	case c.send <- data:
		return true
	default:
		return false
	}
}

// ReadPump 阻塞读取客户端消息，逐帧交给 handler 处理；连接断开时调用 unregister 清理。
// handler 返回 false 时结束读取（如收到 close 指令）。
func (c *Client) ReadPump(unregister func(c *Client), handler func(data []byte) bool) {
	defer func() {
		unregister(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(4096)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		if !handler(data) {
			return
		}
	}
}

// WritePump 阻塞写入：定期发 ping 保活，读取 send 队列下发消息。
// 需在单独 goroutine 中运行，与 ReadPump 并发。
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	// 通知"刚开始心beat"
	for {
		select {
		case data, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
			// 心跳续期回调（在线状态刷新）
			if c.OnlineRefresh != nil {
				c.OnlineRefresh()
			}
		}
	}
}

// CloseGracefully 优雅关闭发送队列（关闭后 WritePump 退出）
func (c *Client) CloseGracefully() {
	close(c.send)
}
