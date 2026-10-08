package ws

import (
	nethttp "net/http"

	"github.com/gorilla/websocket"
)

// Serve 服务一个 WebSocket 连接（网关复用）：
//  1. 从 query 参数 token 解析 JWT（浏览器 WS 无法携带 Authorization header）
//  2. 升级连接
//  3. 创建客户端并注册到 Hub
//  4. 并发启动写循环与读循环，业务帧交由 onMessage 处理
//
// 参数：
//   - secret: JWT 签名密钥（与 AuthServiceMiddleware 一致）
//   - kind: KindUser / KindAgent，决定客户端角色
//   - newClient: 用升级后的连接与 claims.ID 构造 *Client
//   - register/unregister: Hub 注册/注销回调
//   - onMessage: 处理一条业务帧，返回 false 时断开连接
func Serve(w nethttp.ResponseWriter, r *nethttp.Request, secret []byte, kind Kind,
	newClient func(conn *websocket.Conn, id int64) *Client,
	register func(c *Client),
	unregister func(c *Client),
	onMessage func(c *Client, id int64, data []byte) bool) {

	// 鉴权：query token
	claims, err := ParseToken(r.URL.Query().Get("token"), secret)
	if err != nil {
		nethttp.Error(w, "unauthorized", nethttp.StatusUnauthorized)
		return
	}

	// 升级连接
	conn, err := StandardUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// 构造客户端并按角色绑定 ID
	client := newClient(conn, claims.ID)
	if register != nil {
		register(client)
	}

	// 并发读写
	go client.WritePump()
	client.ReadPump(unregister, func(data []byte) bool {
		return onMessage(client, claims.ID, data)
	})
}
