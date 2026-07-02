package blades

import (
	"bufio"
	"bytes"
	"io"
)

// SSEEvent 表示一个 SSE 事件
type SSEEvent struct {
	Event string
	Data  []byte
	ID    string
}

// SSEScanner 用于扫描 SSE 流
type SSEScanner struct {
	scanner *bufio.Scanner
	event   SSEEvent
	lastErr error
}

// NewSSEScanner 创建 SSE 扫描器
func NewSSEScanner(r io.Reader) *SSEScanner {
	return &SSEScanner{
		scanner: bufio.NewScanner(r),
	}
}

// Scan 扫描下一个 SSE 事件
func (s *SSEScanner) Scan() bool {
	s.event = SSEEvent{}
	var dataBuf bytes.Buffer

	for s.scanner.Scan() {
		line := s.scanner.Bytes()
		if len(line) == 0 {
			// 空行表示事件结束
			if dataBuf.Len() > 0 {
				s.event.Data = dataBuf.Bytes()
				return true
			}
			continue
		}

		// 处理 data: 前缀
		if bytes.HasPrefix(line, []byte("data: ")) {
			dataBuf.Write(line[6:])
		} else if bytes.HasPrefix(line, []byte("data:")) {
			dataBuf.Write(line[5:])
		} else if bytes.HasPrefix(line, []byte("event: ")) {
			s.event.Event = string(line[7:])
		} else if bytes.HasPrefix(line, []byte("id: ")) {
			s.event.ID = string(line[4:])
		}
		// 忽略注释行（以:开头）和其他字段
	}

	// 处理最后一段数据（没有尾随空行的情况）
	if dataBuf.Len() > 0 {
		s.event.Data = dataBuf.Bytes()
		return true
	}

	s.lastErr = s.scanner.Err()
	return false
}

// Event 返回当前事件
func (s *SSEScanner) Event() SSEEvent {
	return s.event
}

// Data 返回当前事件的数据
func (s *SSEScanner) Data() []byte {
	return s.event.Data
}

// Err 返回扫描过程中的错误
func (s *SSEScanner) Err() error {
	return s.lastErr
}
