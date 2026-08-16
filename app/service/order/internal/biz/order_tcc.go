package biz

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	v1 "cyi-app/api/service/order/v1"
	pkgdtm "cyi-app/pkg/dtm"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
)

// DtmTCCRepo DTM TCC 订单分支所需的仓储（在 DTM 屏障事务内执行，保证幂等）
type DtmTCCRepo interface {
	// CreateInTx 在 DTM barrier 事务内创建订单（屏障保证同一分支只执行一次）
	CreateInTx(ctx context.Context, tx *sql.Tx, in *Order) error
	// CancelInTx 在 DTM barrier 事务内取消订单：
	// 仅待支付订单可取消；已支付/已取消等状态视为已处理（幂等），订单不存在同样视为已处理
	CancelInTx(ctx context.Context, tx *sql.Tx, orderNo string) error
}

// dtm 分支名常量（与 server 层 pkgdtm.RegisterTCCBranches 注册的分支名保持一致）
const (
	dtmBranchStock = "stock" // 库存分支：预留/释放库存
	dtmBranchOrder = "order" // 订单分支：创建/取消订单
)

// CreateTCC 使用 DTM TCC 编排创建订单：
//   - 分支1（库存）：Try=预留库存，Confirm=空操作（保持预留，支付时才扣减），Cancel=释放库存
//   - 分支2（订单）：Try=创建订单（待支付），Confirm=空操作，Cancel=取消订单
//
// 任一分支 Try 失败时，DTM 自动按序执行已成功分支的 Cancel 补偿，并持久化事务状态、失败自动重试；
// 屏障机制保证分支的幂等、空补偿、悬挂等异常场景。
func (uc *OrderUseCase) CreateTCC(ctx context.Context, in *Order) error {
	if len(in.Items) == 0 {
		return v1.ErrorOrderProductEmpty("order item is empty")
	}

	// 阶段一：校验商品并计算金额（只读，不落库）
	if err := uc.prepareOrder(ctx, in); err != nil {
		return err
	}

	// 确保订单号（order_no 唯一索引，供订单分支回查订单 ID）
	if in.OrderNo == "" {
		in.GenerateOrderNo()
	}
	in.SetStatus(v1.OrderStatus_WaitPay, uc.repo, uc.stockRe, uc.logger)

	// DTM TCC 编排：Try 由本服务进程内发起（用 localURL），Confirm/Cancel 由 DTM 服务回调（用 busiURL）
	gid := strings.ReplaceAll(uuid.NewString(), "-", "")
	local := strings.TrimRight(uc.localURL, "/")
	busi := strings.TrimRight(uc.busiURL, "/")
	err := dtmcli.TccGlobalTransaction2(uc.dtmServer, gid,
		func(tcc *dtmcli.Tcc) {
			// 同步等待全部分支（含失败后的补偿）执行完成，让调用方立即感知结果
			tcc.WaitResult = true
		},
		func(tcc *dtmcli.Tcc) (*resty.Response, error) {
			base := pkgdtm.TCCBasePath
			if _, err := tcc.CallBranch(in,
				local+pkgdtm.BranchURL(base, dtmBranchStock, pkgdtm.OpTry),
				busi+pkgdtm.BranchURL(base, dtmBranchStock, pkgdtm.OpConfirm),
				busi+pkgdtm.BranchURL(base, dtmBranchStock, pkgdtm.OpCancel),
			); err != nil {
				return nil, err
			}
			if _, err := tcc.CallBranch(in,
				local+pkgdtm.BranchURL(base, dtmBranchOrder, pkgdtm.OpTry),
				busi+pkgdtm.BranchURL(base, dtmBranchOrder, pkgdtm.OpConfirm),
				busi+pkgdtm.BranchURL(base, dtmBranchOrder, pkgdtm.OpCancel),
			); err != nil {
				return nil, err
			}
			return nil, nil
		})
	if err != nil {
		uc.log.Errorf("dtm tcc create order failed, gid=%s: %v", gid, err)
		// 将 DTM/resty 原始错误映射为 kratos error，避免 gRPC 透传（形如 rpc error: ...）给前端
		return mapTCCError(err)
	}

	// 通过订单号回查订单 ID（订单号由主流程预生成，订单分支落库）
	created, err := uc.repo.GetOrderByOrderNo(ctx, in.OrderNo)
	if err != nil {
		uc.log.Errorf("query created order by order_no failed, orderNo=%s: %v", in.OrderNo, err)
		return err
	}
	in.ID = created.ID

	// 投递订单超时取消延迟任务（失败不影响主流程，消费端会先校验订单状态保证幂等）
	if err := uc.taskRe.EnqueueTimeoutCancel(ctx, in.ID); err != nil {
		uc.log.Errorf("投递订单超时任务失败, orderID: %d: %v", in.ID, err)
	}
	return nil
}

// DtmStockTry 库存分支 Try：逐 SKU 预留库存（校验可用量后预留）
// 任一 SKU 预留失败时，先内部补偿释放已预留的 SKU，避免 DTM 重试导致重复预留
func (uc *OrderUseCase) DtmStockTry(ctx context.Context, barrier *dtmcli.BranchBarrier, raw []byte) error {
	var in Order
	if err := json.Unmarshal(raw, &in); err != nil {
		return err
	}
	return barrier.CallWithDB(uc.barrierDB, func(tx *sql.Tx) error {
		reserved := make([]int, 0, len(in.Items)) // 已成功预留的 item 下标
		for i, item := range in.Items {
			stock, err := uc.stockRe.GetStock(ctx, item.SkuID)
			if err != nil {
				uc.releaseReservedStock(ctx, in, reserved)
				return err
			}
			if stock.Available < item.Quantity {
				uc.releaseReservedStock(ctx, in, reserved)
				return fmt.Errorf("%w: 商品 SKU %d 库存不足", dtmcli.ErrFailure, item.SkuID)
			}
			if _, err := uc.stockRe.ReserveStock(ctx, item.SkuID, item.Quantity, stock.Version); err != nil {
				uc.releaseReservedStock(ctx, in, reserved)
				// 预留失败（库存不足/版本冲突）属于业务失败，直接终止事务，不做无意义重试
				return fmt.Errorf("%w: 预留库存失败 sku=%d: %v", dtmcli.ErrFailure, item.SkuID, err)
			}
			reserved = append(reserved, i)
		}
		return nil
	})
}

// DtmStockConfirm 库存分支 Confirm：空操作（下单完成保持预留，支付时才真正扣减）
func (uc *OrderUseCase) DtmStockConfirm(ctx context.Context, barrier *dtmcli.BranchBarrier, raw []byte) error {
	return nil
}

// DtmStockCancel 库存分支 Cancel：释放已预留库存（TCC 补偿，幂等：已释放的 SKU 跳过）
func (uc *OrderUseCase) DtmStockCancel(ctx context.Context, barrier *dtmcli.BranchBarrier, raw []byte) error {
	var in Order
	if err := json.Unmarshal(raw, &in); err != nil {
		return err
	}
	return barrier.CallWithDB(uc.barrierDB, func(tx *sql.Tx) error {
		for _, item := range in.Items {
			stock, err := uc.stockRe.GetStock(ctx, item.SkuID)
			if err != nil {
				return err
			}
			// 已无预留（从未预留或已释放）视为已处理，保证幂等
			if stock.Reserved < item.Quantity {
				continue
			}
			if _, err := uc.stockRe.ReleaseStock(ctx, item.SkuID, item.Quantity, stock.Version); err != nil {
				return fmt.Errorf("%w: 释放库存失败 sku=%d: %v", dtmcli.ErrFailure, item.SkuID, err)
			}
		}
		return nil
	})
}

// releaseReservedStock 释放已成功预留的库存（Try 分支失败时的内部补偿，尽力而为）
func (uc *OrderUseCase) releaseReservedStock(ctx context.Context, in Order, reserved []int) {
	for _, i := range reserved {
		item := in.Items[i]
		for retry := 0; retry < 3; retry++ {
			s, e := uc.stockRe.GetStock(ctx, item.SkuID)
			if e != nil {
				uc.log.Errorf("tcc try rollback get stock failed for skuID %d: %v", item.SkuID, e)
				continue
			}
			if _, e = uc.stockRe.ReleaseStock(ctx, item.SkuID, item.Quantity, s.Version); e != nil {
				uc.log.Errorf("tcc try rollback release stock failed for skuID %d (retry %d): %v", item.SkuID, retry, e)
				continue
			}
			uc.log.Infof("tcc try rollback release stock success for skuID %d, qty %d", item.SkuID, item.Quantity)
			break
		}
	}
}

// DtmOrderTry 订单分支 Try：在 DTM 屏障事务内创建订单（待支付）
func (uc *OrderUseCase) DtmOrderTry(ctx context.Context, barrier *dtmcli.BranchBarrier, raw []byte) error {
	var in Order
	if err := json.Unmarshal(raw, &in); err != nil {
		return err
	}
	return barrier.CallWithDB(uc.barrierDB, func(tx *sql.Tx) error {
		return uc.dtmRepo.CreateInTx(ctx, tx, &in)
	})
}

// DtmOrderConfirm 订单分支 Confirm：空操作（订单在 Try 阶段已创建）
func (uc *OrderUseCase) DtmOrderConfirm(ctx context.Context, barrier *dtmcli.BranchBarrier, raw []byte) error {
	return nil
}

// DtmOrderCancel 订单分支 Cancel：取消已创建的待支付订单（TCC 补偿，幂等）
func (uc *OrderUseCase) DtmOrderCancel(ctx context.Context, barrier *dtmcli.BranchBarrier, raw []byte) error {
	var in Order
	if err := json.Unmarshal(raw, &in); err != nil {
		return err
	}
	return barrier.CallWithDB(uc.barrierDB, func(tx *sql.Tx) error {
		return uc.dtmRepo.CancelInTx(ctx, tx, in.OrderNo)
	})
}

// mapTCCError 将 DTM TCC 编排的失败错误映射为用户可感知的 kratos error：
// TCC 分支 Try/Cancel 失败时，dtmcli 返回的是 resty/HTTP 原始错误（含分支回调响应的中文提示），
// 直接返回会给 gRPC 透传成 "rpc error: ..." 前缀展示在前端，因此按失败原因归类为业务错误码
func mapTCCError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "库存不足"):
		return v1.ErrorInsufficientStock("商品库存不足")
	case strings.Contains(msg, "预留库存失败"):
		return v1.ErrorStockOperationFailed("库存预留失败，请稍后重试")
	case strings.Contains(msg, "释放库存失败"):
		return v1.ErrorStockOperationFailed("库存释放失败，请稍后重试")
	default:
		return v1.ErrorStockOperationFailed("下单失败，请稍后重试")
	}
}
