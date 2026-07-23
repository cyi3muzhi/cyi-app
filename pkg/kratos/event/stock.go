package event

const (
	SubjectStockReserve = "stock.reserve"
	SubjectStockDeduct  = "stock.deduct"
	SubjectStockRelease = "stock.release"
	SubjectStockAdd     = "stock.add"
	SubjectStockGet     = "stock.get"
)

// StockReserveRequest 库存操作请求
type StockReserveRequest struct {
	SkuID   int64 `json:"sku_id"`
	Count   int32 `json:"count"`
	Version int32 `json:"version"`
}

// StockOpResponse 库存操作响应
type StockOpResponse struct {
	Success    bool   `json:"success"`
	NewVersion int32  `json:"new_version,omitempty"`
	Available  int32  `json:"available,omitempty"`
	Total      int32  `json:"total,omitempty"`
	Reserved   int32  `json:"reserved,omitempty"`
	Error      string `json:"error,omitempty"`
}
