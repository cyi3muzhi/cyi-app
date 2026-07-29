package biz

import "errors"

// 业务层错误定义（避免直接依赖 proto/conf 等底层包）
var (
	// ErrCreateProductFailed 创建商品失败
	ErrCreateProductFailed = errors.New("create product failed")
	// ErrCreateStockFailed 创建库存失败
	ErrCreateStockFailed = errors.New("create stock failed")
)
