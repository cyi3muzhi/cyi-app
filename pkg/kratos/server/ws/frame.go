package ws

import "encoding/json"

// ClientFrame 客户端 → 服务端的帧信封
// 帧示例：{"type":"send","data":{"conversation_id":1,"msg_type":1,"content":"你好"}}
type ClientFrame struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

// ServerFrame 服务端 → 客户端的帧信封
// 帧示例：{"type":"message","data":{...}}
type ServerFrame struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// NewServerFrame 构造服务端帧并序列化
func NewServerFrame(typ string, data interface{}) []byte {
	payload, err := json.Marshal(&ServerFrame{Type: typ, Data: data})
	if err != nil {
		return nil
	}
	return payload
}
