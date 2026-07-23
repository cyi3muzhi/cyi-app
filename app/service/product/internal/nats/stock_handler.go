package nats

import (
	"context"
	"encoding/json"

	"cyi-app/app/service/product/internal/biz"
	"cyi-app/pkg/kratos/event"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
)

// StockHandler 处理库存相关的NATS消息
type StockHandler struct {
	uc  *biz.StockUseCase
	log *log.Helper
}

// NewStockHandler 创建库存NATS处理器
func NewStockHandler(uc *biz.StockUseCase, logger log.Logger) *StockHandler {
	return &StockHandler{uc: uc, log: log.NewHelper(logger)}
}

// Register 注册NATS订阅
func (h *StockHandler) Register(conn *nats.Conn) error {
	subjects := []string{
		event.SubjectStockReserve,
		event.SubjectStockDeduct,
		event.SubjectStockRelease,
		event.SubjectStockAdd,
		event.SubjectStockGet,
	}
	for _, subject := range subjects {
		_, err := conn.Subscribe(subject, func(msg *nats.Msg) {
			h.log.Infof("received stock event: %s", subject)
			var resp event.StockOpResponse
			switch msg.Subject {
			case event.SubjectStockReserve:
				var req event.StockReserveRequest
				if err := json.Unmarshal(msg.Data, &req); err != nil {
					resp.Error = err.Error()
					break
				}
				newVersion, err := h.uc.ReserveStock(context.Background(), req.SkuID, req.Count, req.Version)
				if err != nil {
					resp.Error = err.Error()
					break
				}
				resp.Success = true
				resp.NewVersion = newVersion
			case event.SubjectStockDeduct:
				var req event.StockReserveRequest
				if err := json.Unmarshal(msg.Data, &req); err != nil {
					resp.Error = err.Error()
					break
				}
				newVersion, err := h.uc.DeductStock(context.Background(), req.SkuID, req.Count, req.Version)
				if err != nil {
					resp.Error = err.Error()
					break
				}
				resp.Success = true
				resp.NewVersion = newVersion
			case event.SubjectStockRelease:
				var req event.StockReserveRequest
				if err := json.Unmarshal(msg.Data, &req); err != nil {
					resp.Error = err.Error()
					break
				}
				newVersion, err := h.uc.ReleaseStock(context.Background(), req.SkuID, req.Count, req.Version)
				if err != nil {
					resp.Error = err.Error()
					break
				}
				resp.Success = true
				resp.NewVersion = newVersion
			case event.SubjectStockAdd:
				var req event.StockReserveRequest
				if err := json.Unmarshal(msg.Data, &req); err != nil {
					resp.Error = err.Error()
					break
				}
				newVersion, err := h.uc.AddStock(context.Background(), req.SkuID, req.Count, req.Version)
				if err != nil {
					resp.Error = err.Error()
					break
				}
				resp.Success = true
				resp.NewVersion = newVersion
			case event.SubjectStockGet:
				var req struct {
					SkuID int64 `json:"sku_id"`
				}
				if err := json.Unmarshal(msg.Data, &req); err != nil {
					resp.Error = err.Error()
					break
				}
				stock, err := h.uc.GetStock(context.Background(), req.SkuID)
				if err != nil {
					resp.Error = err.Error()
					break
				}
				resp.Success = true
				resp.Available = stock.Available
				resp.Total = stock.Total
				resp.Reserved = stock.Reserved
				resp.NewVersion = stock.Version
			}
			// 发送响应
			data, _ := json.Marshal(resp)
			if err := msg.Respond(data); err != nil {
				h.log.Errorf("failed to respond to stock event: %v", err)
			}
		})
		if err != nil {
			return err
		}
	}
	return nil
}
