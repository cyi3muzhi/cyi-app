package nats

import (
	"context"
	"encoding/json"
	"time"

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
			// 使用带超时的 context 调用仓储，避免数据库挂起时无限阻塞；超时后返回错误响应
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var resp event.StockOpResponse
			switch msg.Subject {
			case event.SubjectStockReserve:
				var req event.StockReserveRequest
				if err := json.Unmarshal(msg.Data, &req); err != nil {
					resp.Error = err.Error()
					break
				}
				newVersion, err := h.uc.ReserveStock(ctx, req.SkuID, req.Count, req.Version)
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
				newVersion, err := h.uc.DeductStock(ctx, req.SkuID, req.Count, req.Version)
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
				newVersion, err := h.uc.ReleaseStock(ctx, req.SkuID, req.Count, req.Version)
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
				newVersion, err := h.uc.AddStock(ctx, req.SkuID, req.Count, req.Version)
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
				stock, err := h.uc.GetStock(ctx, req.SkuID)
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
			// 发送响应（序列化失败时记录错误并返回空响应，避免 msg.Respond(nil) 异常）
			data, err := json.Marshal(resp)
			if err != nil {
				h.log.Errorf("failed to marshal stock response: %v", err)
				return
			}
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
