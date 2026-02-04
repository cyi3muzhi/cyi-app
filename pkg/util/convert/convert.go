package convert


import (
	"github.com/mitchellh/mapstructure"
	"github.com/fatih/structs"
)

// 结构体转换
func Convert(src, dst interface{}) error {
    return mapstructure.Decode(src, dst)
}

// 转换为map
func ToMap(s interface{}) map[string]interface{} {
    return structs.Map(s)
}

// 从map转换
// func FromMap(m map[string]interface{}, s interface{}) error {
//     return structs.New()
// }