package convert


import "github.com/mitchellh/mapstructure"

// 结构体转换
func Convert(src, dst interface{}) error {
    return mapstructure.Decode(src, dst)
}