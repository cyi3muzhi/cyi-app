package event

const (
	SubjectStockReserve = "stock.reserve"
	SubjectStockDeduct  = "stock.deduct"
	SubjectStockRelease = "stock.release"
	SubjectStockAdd     = "stock.add"
	SubjectStockGet     = "stock.get"
)

// StockReserveRequest 库存操作请求（reserve/deduct/release/add 共用）
// Gid/BranchID/Op 为 DTM 屏障幂等三元组：携带时库存服务以屏障事务执行，
// 保证跨服务重试场景下同一（gid,branch,op）的库存变更只生效一次（空补偿/防悬挂/幂等）
type StockReserveRequest struct {
	SkuID int64 `json:"sku_id"`
	Count int32 `json:"count"`
	// Gid 全局事务ID（TCC 下单为 tcc-{orderNo}，支付为 pay-{orderNo}；空表示无幂等键，按普通更新执行）
	Gid string `json:"gid,omitempty"`
	// BranchID 分支ID（按 SKU 维度：stock-{skuID}）
	BranchID string `json:"branch_id,omitempty"`
	// Op 屏障操作类型：try / confirm / cancel
	Op string `json:"op,omitempty"`
}

// StockOpResponse 库存操作响应
type StockOpResponse struct {
	Success   bool   `json:"success"`
	Available int32  `json:"available,omitempty"`
	Total     int32  `json:"total,omitempty"`
	Reserved  int32  `json:"reserved,omitempty"`
	Error     string `json:"error,omitempty"`
}
