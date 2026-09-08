package data

import (
	"context"
	"encoding/json"
	"time"

	productRpc "cyi-app/api/service/product/v1"
	"cyi-app/app/service/order/internal/biz"
	"cyi-app/pkg/kratos/event"

	"github.com/go-kratos/kratos/v2/log"
)

// stockNATSTimeout NATS 库存请求超时时间（与 product 订阅侧处理超时对齐，避免对端异常时调用方无限等待）
const stockNATSTimeout = 5 * time.Second

// stockRepo NATS 实现的库存仓储
type stockRepo struct {
	data *Data
	log  *log.Helper
}

// NewStockRepo 创建 NATS 库存仓储
func NewStockRepo(data *Data, logger log.Logger) biz.StockRepo {
	return &stockRepo{data: data, log: log.NewHelper(logger)}
}

// doRequest 发送 NATS 请求并解析响应
// 统一加超时兜底：调用方 ctx 无截止时间或截止时间长于超时时，使用 5s 超时，
// 避免 product 服务/连接异常时订单链路无限阻塞
func (r *stockRepo) doRequest(ctx context.Context, subject string, req interface{}) (*event.StockOpResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, stockNATSTimeout)
	defer cancel()
	msg, err := r.data.nc.RequestWithContext(ctx, subject, data)
	if err != nil {
		return nil, err
	}
	var resp event.StockOpResponse
	if err := json.Unmarshal(msg.Data, &resp); err != nil {
		return nil, err
	}
	if !resp.Success {
		if resp.Error != "" {
			return nil, &stockError{message: resp.Error}
		}
		return nil, &stockError{message: "stock operation failed"}
	}
	return &resp, nil
}

// stockError 库存操作错误
type stockError struct {
	message string
}

func (e *stockError) Error() string {
	return e.message
}

func (r *stockRepo) GetStock(ctx context.Context, skuID int64) (*productRpc.GetStockReply, error) {
	req := struct {
		SkuID int64 `json:"sku_id"`
	}{SkuID: skuID}
	resp, err := r.doRequest(ctx, event.SubjectStockGet, req)
	if err != nil {
		return nil, err
	}
	return &productRpc.GetStockReply{
		Available: resp.Available,
		Total:     resp.Total,
		Reserved:  resp.Reserved,
	}, nil
}

func (r *stockRepo) ReserveStock(ctx context.Context, key biz.StockOpKey, skuID int64, count int32) error {
	req := event.StockReserveRequest{
		SkuID:    skuID,
		Count:    count,
		Gid:      key.Gid,
		BranchID: key.BranchID,
		Op:       key.Op,
	}
	_, err := r.doRequest(ctx, event.SubjectStockReserve, req)
	return err
}

func (r *stockRepo) DeductStock(ctx context.Context, key biz.StockOpKey, skuID int64, count int32) error {
	req := event.StockReserveRequest{
		SkuID:    skuID,
		Count:    count,
		Gid:      key.Gid,
		BranchID: key.BranchID,
		Op:       key.Op,
	}
	_, err := r.doRequest(ctx, event.SubjectStockDeduct, req)
	return err
}

func (r *stockRepo) ReleaseStock(ctx context.Context, key biz.StockOpKey, skuID int64, count int32) error {
	req := event.StockReserveRequest{
		SkuID:    skuID,
		Count:    count,
		Gid:      key.Gid,
		BranchID: key.BranchID,
		Op:       key.Op,
	}
	_, err := r.doRequest(ctx, event.SubjectStockRelease, req)
	return err
}

func (r *stockRepo) AddStock(ctx context.Context, key biz.StockOpKey, skuID int64, count int32) error {
	req := event.StockReserveRequest{
		SkuID:    skuID,
		Count:    count,
		Gid:      key.Gid,
		BranchID: key.BranchID,
		Op:       key.Op,
	}
	_, err := r.doRequest(ctx, event.SubjectStockAdd, req)
	return err
}
