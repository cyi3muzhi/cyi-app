// Package ws 提供 WebSocket 连接管理与实时消息分发（Hub）工具。
// 支持用户端（interface 网关）与客服端（admin 网关）两类角色，
// 由各网关持有 Hub 实例并订阅 NATS chat.message，按目标 ID 推送至本地连接。
package ws

import (
	nethttp "net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Kind 连接角色类型
type Kind int

const (
	// KindUser 用户连接（interface 网关）
	KindUser Kind = 1
	// KindAgent 客服连接（admin 网关）
	KindAgent Kind = 2
)

// Hub WebSocket 连接分发器
// 用户/客服可多端登录，map 内为多个连接；按 ID 推送时遍历该用户/客服的所有连接。
type Hub struct {
	mu     sync.RWMutex
	users  map[int64]map[*Client]struct{} // userID -> connections
	agents map[int64]map[*Client]struct{} // agentID -> connections
}

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		users:  make(map[int64]map[*Client]struct{}),
		agents: make(map[int64]map[*Client]struct{}),
	}
}

// Register 注册连接（按角色放入对应 map）
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set := h.setFor(c, true)
	set[c] = struct{}{}
}

// Unregister 注销连接
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if set := h.setFor(c, false); set != nil {
		delete(set, c)
	}
}

// setFor 返回连接所属的连接集合；create 为 true 时若不存在则创建
func (h *Hub) setFor(c *Client, create bool) map[*Client]struct{} {
	key := c.userID
	m := h.users
	if c.kind == KindAgent {
		key = c.agentID
		m = h.agents
	}
	set, ok := m[key]
	if !ok && create {
		set = make(map[*Client]struct{})
		m[key] = set
	}
	return set
}

// PushToUser 推送消息给指定用户的所有连接
func (h *Hub) PushToUser(userID int64, data []byte) {
	h.push(KindUser, userID, data)
}

// PushToAgent 推送消息给指定客服的所有连接
func (h *Hub) PushToAgent(agentID int64, data []byte) {
	h.push(KindAgent, agentID, data)
}

// push 按角色与 ID 遍历连接推送（写入各自的非阻塞 channel）
func (h *Hub) push(kind Kind, id int64, data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var set map[*Client]struct{}
	if kind == KindAgent {
		set = h.agents[id]
	} else {
		set = h.users[id]
	}
	for c := range set {
		c.Send(data)
	}
}

// BroadcastUsers 广播给所有用户连接（M3 会话状态事件可选，预留）
func (h *Hub) BroadcastUsers(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, set := range h.users {
		for c := range set {
			c.Send(data)
		}
	}
}

// BroadcastAgents 广播给所有客服连接（M3 待接入列表刷新预留）
func (h *Hub) BroadcastAgents(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, set := range h.agents {
		for c := range set {
			c.Send(data)
		}
	}
}

// StandardUpgrader 通用 WebSocket 升级器（生产环境应限制 Origin）
var StandardUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *nethttp.Request) bool {
		return true
	},
}
