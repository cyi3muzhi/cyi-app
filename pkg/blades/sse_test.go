package blades

import (
	"strings"
	"testing"
)

// TestSSEScanner_SingleEvent 测试单个SSE事件解析
func TestSSEScanner_SingleEvent(t *testing.T) {
	input := "data: hello world\n\n"
	scanner := NewSSEScanner(strings.NewReader(input))

	if !scanner.Scan() {
		t.Fatal("expected scan to return true for first event")
	}

	if string(scanner.Data()) != "hello world" {
		t.Errorf("expected data 'hello world', got '%s'", string(scanner.Data()))
	}

	if scanner.Scan() {
		t.Error("expected scan to return false after last event")
	}

	if scanner.Err() != nil {
		t.Errorf("unexpected error: %v", scanner.Err())
	}
}

// TestSSEScanner_MultipleEvents 测试多个SSE事件解析
func TestSSEScanner_MultipleEvents(t *testing.T) {
	input := "data: first event\n\n" +
		"data: second event\n\n" +
		"data: third event\n\n"

	scanner := NewSSEScanner(strings.NewReader(input))

	expected := []string{"first event", "second event", "third event"}
	var actual []string

	for scanner.Scan() {
		actual = append(actual, string(scanner.Data()))
	}

	if len(actual) != len(expected) {
		t.Fatalf("expected %d events, got %d", len(expected), len(actual))
	}

	for i, exp := range expected {
		if actual[i] != exp {
			t.Errorf("event %d: expected '%s', got '%s'", i, exp, actual[i])
		}
	}
}

// TestSSEScanner_DataWithoutSpace 测试data:后没有空格的情况
func TestSSEScanner_DataWithoutSpace(t *testing.T) {
	input := "data:hello\n\n"
	scanner := NewSSEScanner(strings.NewReader(input))

	if !scanner.Scan() {
		t.Fatal("expected scan to return true")
	}

	if string(scanner.Data()) != "hello" {
		t.Errorf("expected data 'hello', got '%s'", string(scanner.Data()))
	}
}

// TestSSEScanner_MultilineData 测试多行data合并
func TestSSEScanner_MultilineData(t *testing.T) {
	input := "data: line1\n" +
		"data: line2\n" +
		"data: line3\n\n"

	scanner := NewSSEScanner(strings.NewReader(input))

	if !scanner.Scan() {
		t.Fatal("expected scan to return true")
	}

	expected := "line1line2line3"
	if string(scanner.Data()) != expected {
		t.Errorf("expected data '%s', got '%s'", expected, string(scanner.Data()))
	}
}

// TestSSEScanner_WithEventField 测试带event字段的SSE事件
func TestSSEScanner_WithEventField(t *testing.T) {
	input := "event: message\n" +
		"data: hello\n\n"

	scanner := NewSSEScanner(strings.NewReader(input))

	if !scanner.Scan() {
		t.Fatal("expected scan to return true")
	}

	event := scanner.Event()
	if event.Event != "message" {
		t.Errorf("expected event 'message', got '%s'", event.Event)
	}
	if string(event.Data) != "hello" {
		t.Errorf("expected data 'hello', got '%s'", string(event.Data))
	}
}

// TestSSEScanner_WithIDField 测试带id字段的SSE事件
func TestSSEScanner_WithIDField(t *testing.T) {
	input := "id: 12345\n" +
		"data: test data\n\n"

	scanner := NewSSEScanner(strings.NewReader(input))

	if !scanner.Scan() {
		t.Fatal("expected scan to return true")
	}

	event := scanner.Event()
	if event.ID != "12345" {
		t.Errorf("expected id '12345', got '%s'", event.ID)
	}
	if string(event.Data) != "test data" {
		t.Errorf("expected data 'test data', got '%s'", string(event.Data))
	}
}

// TestSSEScanner_AllFields 测试包含所有字段的SSE事件
func TestSSEScanner_AllFields(t *testing.T) {
	input := "id: 1\n" +
		"event: update\n" +
		"data: payload\n\n"

	scanner := NewSSEScanner(strings.NewReader(input))

	if !scanner.Scan() {
		t.Fatal("expected scan to return true")
	}

	event := scanner.Event()
	if event.ID != "1" {
		t.Errorf("expected id '1', got '%s'", event.ID)
	}
	if event.Event != "update" {
		t.Errorf("expected event 'update', got '%s'", event.Event)
	}
	if string(event.Data) != "payload" {
		t.Errorf("expected data 'payload', got '%s'", string(event.Data))
	}
}

// TestSSEScanner_NoTrailingNewline 测试没有尾部空行的情况
func TestSSEScanner_NoTrailingNewline(t *testing.T) {
	input := "data: final event"
	scanner := NewSSEScanner(strings.NewReader(input))

	if !scanner.Scan() {
		t.Fatal("expected scan to return true for final event without trailing newline")
	}

	if string(scanner.Data()) != "final event" {
		t.Errorf("expected data 'final event', got '%s'", string(scanner.Data()))
	}

	if scanner.Scan() {
		t.Error("expected scan to return false after final event")
	}
}

// TestSSEScanner_EmptyInput 测试空输入
func TestSSEScanner_EmptyInput(t *testing.T) {
	input := ""
	scanner := NewSSEScanner(strings.NewReader(input))

	if scanner.Scan() {
		t.Error("expected scan to return false for empty input")
	}
}

// TestSSEScanner_OnlyEmptyLines 测试只有空行的输入
func TestSSEScanner_OnlyEmptyLines(t *testing.T) {
	input := "\n\n\n"
	scanner := NewSSEScanner(strings.NewReader(input))

	if scanner.Scan() {
		t.Error("expected scan to return false for only empty lines")
	}
}

// TestSSEScanner_CommentLines 测试注释行（以:开头）应该被忽略
func TestSSEScanner_CommentLines(t *testing.T) {
	input := ": this is a comment\n" +
		"data: actual data\n\n"

	scanner := NewSSEScanner(strings.NewReader(input))

	if !scanner.Scan() {
		t.Fatal("expected scan to return true")
	}

	if string(scanner.Data()) != "actual data" {
		t.Errorf("expected data 'actual data', got '%s'", string(scanner.Data()))
	}
}

// TestSSEScanner_EmptyDataEvent 测试data为空的事件不被计数
func TestSSEScanner_EmptyDataEvent(t *testing.T) {
	input := "event: keepalive\n\n" +
		"data: real event\n\n"

	scanner := NewSSEScanner(strings.NewReader(input))

	events := 0
	for scanner.Scan() {
		events++
	}

	if events != 1 {
		t.Errorf("expected 1 event with data, got %d", events)
	}
}

// TestSSEScanner_DataMethod 测试Data()方法返回数据
func TestSSEScanner_DataMethod(t *testing.T) {
	input := "data: test123\n\n"
	scanner := NewSSEScanner(strings.NewReader(input))

	scanner.Scan()
	data := scanner.Data()

	if string(data) != "test123" {
		t.Errorf("Data() expected 'test123', got '%s'", string(data))
	}
}

// TestSSEScanner_EventMethod 测试Event()方法返回完整事件
func TestSSEScanner_EventMethod(t *testing.T) {
	input := "id: 42\n" +
		"event: custom\n" +
		"data: hello\n\n"

	scanner := NewSSEScanner(strings.NewReader(input))
	scanner.Scan()
	event := scanner.Event()

	if event.ID != "42" {
		t.Errorf("Event().ID expected '42', got '%s'", event.ID)
	}
	if event.Event != "custom" {
		t.Errorf("Event().Event expected 'custom', got '%s'", event.Event)
	}
	if string(event.Data) != "hello" {
		t.Errorf("Event().Data expected 'hello', got '%s'", string(event.Data))
	}
}

// TestSSEScanner_ErrMethod 测试Err()方法
func TestSSEScanner_ErrMethod(t *testing.T) {
	input := "data: test\n\n"
	scanner := NewSSEScanner(strings.NewReader(input))

	for scanner.Scan() {
	}

	if scanner.Err() != nil {
		t.Errorf("expected no error, got: %v", scanner.Err())
	}
}

// TestSSEScanner_DeepSeekStreamFormat 测试DeepSeek风格的SSE流格式
func TestSSEScanner_DeepSeekStreamFormat(t *testing.T) {
	input := "data: {\"choices\":[{\"delta\":{\"content\":\"Hello\"}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"content\":\" World\"}}]}\n\n" +
		"data: [DONE]\n\n"

	scanner := NewSSEScanner(strings.NewReader(input))

	var events []string
	for scanner.Scan() {
		events = append(events, string(scanner.Data()))
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if !strings.Contains(events[0], "Hello") {
		t.Errorf("first event should contain 'Hello', got: %s", events[0])
	}
	if !strings.Contains(events[1], " World") {
		t.Errorf("second event should contain ' World', got: %s", events[1])
	}
	if events[2] != "[DONE]" {
		t.Errorf("third event should be '[DONE]', got: %s", events[2])
	}
}
