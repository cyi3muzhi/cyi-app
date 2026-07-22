package data

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMarshalSpec_Empty 测试空规格 map 序列化应返回空字符串
func TestMarshalSpec_Empty(t *testing.T) {
	assert.Equal(t, "", marshalSpec(nil))
	assert.Equal(t, "", marshalSpec(map[string]string{}))
}

// TestMarshalSpec_Normal 测试正常规格 map 序列化为 JSON 字符串
func TestMarshalSpec_Normal(t *testing.T) {
	input := map[string]string{"颜色": "红色", "尺码": "XL"}
	result := marshalSpec(input)
	// map 的 JSON 键顺序不确定，通过反序列化验证内容
	decoded := unmarshalSpec(result)
	assert.Equal(t, input, decoded)
}

// TestUnmarshalSpec_Empty 测试空字符串反序列化应返回 nil
func TestUnmarshalSpec_Empty(t *testing.T) {
	assert.Nil(t, unmarshalSpec(""))
}

// TestUnmarshalSpec_Normal 测试正常 JSON 字符串反序列化为规格 map
func TestUnmarshalSpec_Normal(t *testing.T) {
	input := `{"颜色":"红色","尺码":"XL"}`
	result := unmarshalSpec(input)
	assert.Equal(t, map[string]string{"颜色": "红色", "尺码": "XL"}, result)
}

// TestUnmarshalSpec_InvalidJSON 测试非法 JSON 应安全返回 nil 而非 panic
func TestUnmarshalSpec_InvalidJSON(t *testing.T) {
	result := unmarshalSpec("{invalid json}")
	assert.Nil(t, result)
}

// TestMarshalUnmarshalRoundTrip 测试序列化与反序列化的往返一致性
func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	original := map[string]string{"颜色": "红色", "尺码": "XL"}
	encoded := marshalSpec(original)
	decoded := unmarshalSpec(encoded)
	assert.Equal(t, original, decoded)
}
