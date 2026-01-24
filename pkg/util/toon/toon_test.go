package toon

import (
	"strings"
	"testing"
)

// 测试基本结构体
 type BasicStruct struct {
	Name  string `toon:"name"`
	Age   int    `toon:"age"`
	Score float64
	Active bool
}

// 测试嵌套结构体
type NestedStruct struct {
	Basic BasicStruct
	Tags  []string
}

// 测试包含映射的结构体
type MapStruct struct {
	Config map[string]interface{}
}

// TestMarshalStruct 测试结构体序列化
func TestMarshalStruct(t *testing.T) {
	// 测试基本结构体
	basic := BasicStruct{
		Name:   "test",
		Age:    18,
		Score:  95.5,
		Active: true,
	}

	result, err := Marshal(basic)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := `name test
age 18
Score 95.5
Active true
`

	if result != expected {
		t.Errorf("MarshalStruct mismatch. Expected:\n%q, got:\n%q", expected, result)
	}

	// 测试嵌套结构体
	nested := NestedStruct{
		Basic: BasicStruct{
			Name:   "nested",
			Age:    20,
			Score:  88.8,
			Active: false,
		},
		Tags: []string{"tag1", "tag2"},
	}

	result, err = Marshal(nested)
	if err != nil {
		t.Fatalf("Marshal nested struct failed: %v", err)
	}

	expected = `Basic 
  name nested
  age 20
  Score 88.8
  Active false
Tags 
  tag1
  tag2
`

	if result != expected {
		t.Errorf("MarshalNestedStruct mismatch. Expected:\n%q, got:\n%q", expected, result)
	}
}

// TestUnmarshalStruct 测试结构体反序列化
func TestUnmarshalStruct(t *testing.T) {
	// 测试基本结构体
	toonData := `name test
age 18
Score 95.5
Active true
`

	var basic BasicStruct
	err := Unmarshal(toonData, &basic)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if basic.Name != "test" || basic.Age != 18 || basic.Score != 95.5 || basic.Active != true {
		t.Errorf("UnmarshalStruct mismatch. Got: %+v", basic)
	}

	// 测试嵌套结构体
	nestedData := `Basic 
  name nested
  age 20
  Score 88.8
  Active false
Tags 
  tag1
  tag2
`

	var nested NestedStruct
	err = Unmarshal(nestedData, &nested)
	if err != nil {
		t.Fatalf("Unmarshal nested struct failed: %v", err)
	}

	if nested.Basic.Name != "nested" || nested.Basic.Age != 20 || len(nested.Tags) != 2 {
		t.Errorf("UnmarshalNestedStruct mismatch. Got: %+v", nested)
	}
}

// TestStructWithMap 测试包含映射的结构体
func TestStructWithMap(t *testing.T) {
	// 测试序列化
	mapStruct := MapStruct{
		Config: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
			"key3": true,
		},
	}

	result, err := Marshal(mapStruct)
	if err != nil {
		t.Fatalf("Marshal struct with map failed: %v", err)
	}
	t.Logf("\n%v",result)

	// 检查输出是否包含所有预期的键值对，不依赖顺序
	expectedKeys := []string{"Config", "key1", "value1", "key2", "123", "key3", "true"}
	for _, key := range expectedKeys {
		if !strings.Contains(result, key) {
			t.Errorf("MarshalStructWithMap missing key: %s, got: %q", key, result)
		}
	}

	// 测试反序列化
	toonData := `Config 
  key1 value1
  key2 123
  key3 true
`

	var unmarshalMapStruct MapStruct
	err = Unmarshal(toonData, &unmarshalMapStruct)
	if err != nil {
		t.Fatalf("Unmarshal struct with map failed: %v", err)
	}
	t.Logf("unmarshalMapStruct: %v",unmarshalMapStruct)

	if len(unmarshalMapStruct.Config) != 3 {
		t.Errorf("UnmarshalStructWithMap mismatch. Expected 3 map entries, got %d", len(unmarshalMapStruct.Config))
	}
}

// TestEmptyStruct 测试空结构体
func TestEmptyStruct(t *testing.T) {
	type EmptyStruct struct {}

	empty := EmptyStruct{}
	result, err := Marshal(empty)
	if err != nil {
		t.Fatalf("Marshal empty struct failed: %v", err)
	}

	// 空结构体应该生成空字符串
	if result != "" {
		t.Errorf("MarshalEmptyStruct mismatch. Expected empty string, got %q", result)
	}

	// 测试反序列化到空结构体（使用基本 TOON 数据）
	var unmarshalEmpty EmptyStruct
	err = Unmarshal("name test\nage 18", &unmarshalEmpty)
	if err != nil {
		t.Fatalf("Unmarshal empty struct failed: %v", err)
	}
}

// TestBasicTypes 测试基本类型序列化和反序列化
func TestBasicTypes(t *testing.T) {
	// 测试数字
	numData := "123.45"
	numNode, err := Parse(numData)
	if err != nil {
		t.Fatalf("Parse number failed: %v", err)
	}
	if numNode.Value() != 123.45 {
		t.Errorf("Number value mismatch. Expected: 123.45, got: %v", numNode.Value())
	}

	// 测试布尔值
	boolData := "true"
	boolNode, err := Parse(boolData)
	if err != nil {
		t.Fatalf("Parse boolean failed: %v", err)
	}
	if boolNode.Value() != true {
		t.Errorf("Boolean value mismatch. Expected: true, got: %v", boolNode.Value())
	}

	// 测试单字符串（不含空格）
	singleStrData := "test"
	singleStrNode, err := Parse(singleStrData)
	if err != nil {
		t.Fatalf("Parse single string failed: %v", err)
	}
	if singleStrNode.Value() != singleStrData {
		t.Errorf("Single string value mismatch. Expected: %q, got: %v", singleStrData, singleStrNode.Value())
	}
}

// TestObject 测试对象类型
func TestObject(t *testing.T) {
	// 测试反序列化
	toonData := `name test
age 18
active true
`

	var unmarshalObj map[string]interface{}
	err := Unmarshal(toonData, &unmarshalObj)
	if err != nil {
		t.Fatalf("Unmarshal object failed: %v", err)
	}

	if len(unmarshalObj) != 3 {
		t.Errorf("UnmarshalObject mismatch. Expected 3 keys, got %d", len(unmarshalObj))
	}

	// 测试使用 map[string]interface{} 进行序列化
	objMap := map[string]interface{}{
		"name":   "test",
		"age":    18,
		"active": true,
	}

	result, err := Marshal(objMap)
	if err != nil {
		t.Fatalf("Marshal object map failed: %v", err)
	}

	// 检查输出是否包含所有预期的键值对
	expectedKeys := []string{"name", "test", "age", "18", "active", "true"}
	for _, key := range expectedKeys {
		if !strings.Contains(result, key) {
			t.Errorf("MarshalObject missing key: %s, got: %q", key, result)
		}
	}
}

// TestArray 测试数组类型
func TestArray(t *testing.T) {
	// 测试使用 []interface{} 进行序列化
	arr := []interface{}{"item1", 2, false}

	result, err := Marshal(arr)
	if err != nil {
		t.Fatalf("Marshal array failed: %v", err)
	}

	// 检查输出是否包含所有预期项
	expectedItems := []string{"item1", "2", "false"}
	for _, item := range expectedItems {
		if !strings.Contains(result, item) {
			t.Errorf("MarshalArray missing item: %s, got: %q", item, result)
		}
	}

	// 测试使用数组字段的结构体进行反序列化
	type ArrayStruct struct {
		Items []string
	}

	toonData := `Items 
  item1
  item2
  item3
`

	var unmarshalStruct ArrayStruct
	err = Unmarshal(toonData, &unmarshalStruct)
	if err != nil {
		t.Fatalf("Unmarshal array struct failed: %v", err)
	}

	if len(unmarshalStruct.Items) != 3 {
		t.Errorf("UnmarshalArrayStruct mismatch. Expected 3 items, got %d", len(unmarshalStruct.Items))
	}
}
