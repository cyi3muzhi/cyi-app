//go:build integration

// Package integration 是订单下单链路（DTM TCC 编排 + NATS 库存操作 + DTM 屏障幂等）的集成测试。
//
// 运行前置条件（仅启动基础设施，不启动 product/order 业务服务——product 侧库存逻辑由测试替身模拟）：
//
//	docker compose -f deploy/docker-compose/docker-compose.infra.yml up -d mysql nats dtm
//
// 运行方式（PowerShell）：
//
//	$env:CYC_TCC_INTEGRATION="1"; go test -tags=integration -v ./app/service/order/integration/...
//
// 可覆盖环境变量：MYSQL_DSN（默认 root:mysql0411@tcp(localhost:3306)/cin_test）、
// NATS_URL（默认 nats://localhost:4222）、DTM_SERVER（默认 http://localhost:36789/api/dtmsvr）、
// DTM_BUSI_HOST（默认 host.docker.internal，DTM 容器回调宿主机分支 HTTP 服务用）。
package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	productRpc "cyi-app/api/service/product/v1"
	v1 "cyi-app/api/service/order/v1"
	"cyi-app/app/service/order/internal/biz"
	"cyi-app/app/service/order/internal/data"
	mod "cyi-app/app/service/order/internal/data/models"
	pkgdtm "cyi-app/pkg/dtm"
	"cyi-app/pkg/kratos/event"

	"github.com/dtm-labs/dtm/client/dtmcli"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// stockDDL 测试替身使用的库存表（结构与 product 服务 stock 表核心列对齐；销量列在 sku/product 表，替身不维护）
const stockDDL = `
CREATE TABLE IF NOT EXISTS stock (
  id          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  product_id  bigint NOT NULL DEFAULT 0 COMMENT '商品ID',
  sku_id      bigint NOT NULL DEFAULT 0 COMMENT 'SKU ID',
  total       bigint NOT NULL DEFAULT 0 COMMENT '总库存',
  available   bigint NOT NULL DEFAULT 0 COMMENT '可用库存',
  reserved    bigint NOT NULL DEFAULT 0 COMMENT '预留库存',
  created_at  datetime(3) DEFAULT NULL,
  updated_at  datetime(3) DEFAULT NULL,
  deleted_at  datetime(3) DEFAULT NULL,
  PRIMARY KEY (id),
  KEY idx_stock_sku_id (sku_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='TCC 集成测试库存表（product 服务替身）'`

// harness 集成测试夹具：真实 MySQL/NATS/DTM + 真实 order data/biz 组件 + product 服务 NATS 替身
type harness struct {
	db     *gorm.DB    // order 业务库连接（gorm，AutoMigrate 后使用）
	sqlDB  *sql.DB     // 原生连接（DTM 屏障与库存替身 SQL 使用）
	nc     *nats.Conn  // NATS 连接
	uc     *biz.OrderUseCase
	repo   biz.OrderRepo
	userID    int64 // 每次运行唯一，避免历史数据干扰
	productID int64
}

// envOr 读取环境变量，未设置时返回默认值
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// newHarness 组装集成测试夹具：连接 MySQL/NATS、建表、启动 product 替身订阅与 DTM 分支 HTTP 服务
func newHarness(t *testing.T) *harness {
	t.Helper()
	if os.Getenv("CYC_TCC_INTEGRATION") != "1" {
		t.Skip("跳过 TCC 集成测试：设置环境变量 CYC_TCC_INTEGRATION=1 并启动 mysql/nats/dtm 基础设施后运行")
	}

	dsn := envOr("MYSQL_DSN", "root:mysql0411@tcp(localhost:3306)/cin_test?charset=utf8mb4&parseTime=True&loc=Local")
	natsURL := envOr("NATS_URL", "nats://localhost:4222")
	dtmServer := envOr("DTM_SERVER", "http://localhost:36789/api/dtmsvr")
	busiHost := envOr("DTM_BUSI_HOST", "host.docker.internal")

	// 1. 连接 MySQL 并准备表结构
	gdb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy:         schema.NamingStrategy{SingularTable: true},
		SkipDefaultTransaction: true,
	})
	if err != nil {
		t.Fatalf("连接 MySQL 失败: %v", err)
	}
	if err := gdb.AutoMigrate(&mod.Order{}, &mod.OrderItem{}, &mod.OrderLogistics{}, &mod.PaymentLog{}, &mod.OrderRefund{}); err != nil {
		t.Fatalf("订单表 AutoMigrate 失败: %v", err)
	}
	if err := gdb.Exec(stockDDL).Error; err != nil {
		t.Fatalf("创建库存替身表失败: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("获取原生 DB 连接失败: %v", err)
	}

	// 2. 连接 NATS 并注册 product 服务替身（注意：运行时不要同时启动 product 服务，避免订阅者竞争）
	nc, err := nats.Connect(natsURL)
	if err != nil {
		t.Fatalf("连接 NATS 失败: %v", err)
	}
	fp := &fakeProduct{db: sqlDB}
	fp.start(t, nc)

	// 3. 组装 order 真实组件（data 层 NewData 会设置 curd 全局默认 DB）
	d, cleanup, err := data.NewData(nil, gdb, log.DefaultLogger, nil, nil, nc, nil)
	if err != nil {
		t.Fatalf("初始化 data.Data 失败: %v", err)
	}
	t.Cleanup(cleanup)
	orderRepo := data.NewOrderRepo(d, log.DefaultLogger)
	stockRepo := data.NewStockRepo(d, log.DefaultLogger)
	dtmStore := data.NewDtmOrderStore(nil, log.DefaultLogger)

	h := &harness{
		db:        gdb,
		sqlDB:     sqlDB,
		nc:        nc,
		repo:      orderRepo,
		userID:    time.Now().UnixNano()%100000 + 700000,
		productID: time.Now().UnixNano()/1e6%100000 + 800000,
	}

	// 4. 启动 DTM 分支 HTTP 服务（Try 进程内自调用用 localhost，Confirm/Cancel 由 DTM 容器经 host.docker.internal 回调）
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("监听分支 HTTP 端口失败: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	port := ln.Addr().(*net.TCPAddr).Port
	localURL := fmt.Sprintf("http://localhost:%d", port)
	busiURL := fmt.Sprintf("http://%s:%d", busiHost, port)

	h.uc = biz.NewOrderUseCase(
		orderRepo,
		&stubProductRepo{productID: h.productID},
		stockRepo,
		&stubTaskRepo{},
		dtmStore,
		sqlDB,
		pkgdtm.ServerAddr(dtmServer),
		pkgdtm.BusiURL(busiURL),
		pkgdtm.LocalURL(localURL),
		log.DefaultLogger,
	)

	mux := http.NewServeMux()
	reg := func(branch, op string, fn pkgdtm.BranchFunc) {
		mux.HandleFunc(pkgdtm.BranchURL(pkgdtm.TCCBasePath, branch, op), pkgdtm.BranchHandler(fn))
	}
	reg("stock", pkgdtm.OpTry, h.uc.DtmStockTry)
	reg("stock", pkgdtm.OpConfirm, h.uc.DtmStockConfirm)
	reg("stock", pkgdtm.OpCancel, h.uc.DtmStockCancel)
	reg("order", pkgdtm.OpTry, h.uc.DtmOrderTry)
	reg("order", pkgdtm.OpConfirm, h.uc.DtmOrderConfirm)
	reg("order", pkgdtm.OpCancel, h.uc.DtmOrderCancel)
	go func() { _ = http.Serve(ln, mux) }()

	return h
}

// seedStock 准备一个 SKU 的库存数据（先清理同 SKU 历史行，保证重复运行结果稳定）
func (h *harness) seedStock(t *testing.T, skuID int64, total int32) {
	t.Helper()
	if err := h.db.Exec("DELETE FROM stock WHERE sku_id = ?", skuID).Error; err != nil {
		t.Fatalf("清理库存历史数据失败: %v", err)
	}
	if err := h.db.Exec(
		"INSERT INTO stock (product_id, sku_id, total, available, reserved, created_at, updated_at) VALUES (?, ?, ?, ?, ?, NOW(3), NOW(3))",
		h.productID, skuID, total, total, 0,
	).Error; err != nil {
		t.Fatalf("写入库存初始数据失败: %v", err)
	}
}

// stockSnap 查询指定 SKU 的库存三元组（total/available/reserved）
func (h *harness) stockSnap(t *testing.T, skuID int64) (total, available, reserved int32) {
	t.Helper()
	if err := h.sqlDB.QueryRow(
		"SELECT total, available, reserved FROM stock WHERE sku_id = ? AND deleted_at IS NULL", skuID,
	).Scan(&total, &available, &reserved); err != nil {
		t.Fatalf("查询库存快照失败 skuID=%d: %v", skuID, err)
	}
	return
}

// uniqueSkuID 生成每次运行唯一的 SKU ID，避免与历史数据/屏障记录冲突
func uniqueSkuID() int64 {
	return time.Now().UnixNano()/1e6%100000 + 900000
}

// newOrderInput 构造下单入参（商品单价 100 元，数量由调用方指定）
func (h *harness) newOrderInput(idempotencyKey string, skuID int64, quantity int32) *biz.Order {
	return &biz.Order{
		UserID:         h.userID,
		IdempotencyKey: idempotencyKey,
		Items: []*biz.OrderItem{
			{ProductID: h.productID, SkuID: skuID, Quantity: quantity, Currency: "CNY"},
		},
	}
}

// stubProductRepo 商品服务桩：固定返回上架商品（单价 10000 分=100 元），供 prepareOrder 校验与算价
type stubProductRepo struct {
	productID int64
}

// GetProduct 返回固定的上架商品信息
func (s *stubProductRepo) GetProduct(_ context.Context, productID int64) (*productRpc.GetProductReply, error) {
	return &productRpc.GetProductReply{
		Product: &productRpc.ProductItem{
			ProductId: productID,
			Name:      "集成测试商品",
			Price:     10000,
			Status:    1,
			Image:     "http://example.com/p.png",
		},
	}, nil
}

// stubTaskRepo 订单超时任务桩：集成测试不投递 asynq 延迟任务
type stubTaskRepo struct{}

// EnqueueTimeoutCancel 空实现（超时取消由 cron/人工兜底，不在本测试范围）
func (*stubTaskRepo) EnqueueTimeoutCancel(_ context.Context, _ int64) error { return nil }

// fakeProduct product 服务 NATS 替身：订阅库存 5 个 subject，
// 在 DTM 屏障事务内执行与 product 服务 internal/data/stock.go 相同语义的库存 SQL（不维护销量表）
type fakeProduct struct {
	db *sql.DB
}

// start 注册 5 个库存 subject 的订阅，测试结束时自动退订
func (f *fakeProduct) start(t *testing.T, nc *nats.Conn) {
	t.Helper()
	// stockOp 屏障/普通事务内执行一条库存 SQL，返回响应体
	type stockOp func(ctx context.Context, tx *sql.Tx, req *event.StockReserveRequest) (event.StockOpResponse, error)
	subscribe := func(subject string, op stockOp) {
		s, err := nc.Subscribe(subject, func(msg *nats.Msg) {
			var req event.StockReserveRequest
			if err := json.Unmarshal(msg.Data, &req); err != nil {
				f.respond(msg, event.StockOpResponse{Success: false, Error: "bad request: " + err.Error()})
				return
			}
			var resp event.StockOpResponse
			var err error
			// 携带屏障三元组时走 DTM 屏障事务（与 product 服务 runWithBarrier 一致），否则普通事务
			if req.Gid != "" {
				barrier := &dtmcli.BranchBarrier{TransType: "tcc", Gid: req.Gid, BranchID: req.BranchID, Op: req.Op}
				err = barrier.CallWithDB(f.db, func(tx *sql.Tx) error {
					r, e := op(context.Background(), tx, &req)
					resp = r
					return e
				})
			} else {
				var tx *sql.Tx
				tx, err = f.db.BeginTx(context.Background(), nil)
				if err == nil {
					resp, err = op(context.Background(), tx, &req)
					if err != nil {
						_ = tx.Rollback()
					} else {
						err = tx.Commit()
					}
				}
			}
			if err != nil {
				f.respond(msg, event.StockOpResponse{Success: false, Error: err.Error()})
				return
			}
			resp.Success = true
			f.respond(msg, resp)
		})
		if err != nil {
			t.Fatalf("订阅 %s 失败: %v", subject, err)
		}
		t.Cleanup(func() { _ = s.Drain() })
	}
	subscribe(event.SubjectStockReserve, f.reserve)
	subscribe(event.SubjectStockDeduct, f.deduct)
	subscribe(event.SubjectStockRelease, f.release)
	subscribe(event.SubjectStockAdd, f.add)
	subscribe(event.SubjectStockGet, f.get)
}

// respond 回复 NATS 请求消息
func (f *fakeProduct) respond(msg *nats.Msg, resp event.StockOpResponse) {
	data, _ := json.Marshal(resp)
	_ = msg.Respond(data)
}

// reserve 预留库存（available-，reserved+），可用量不足返回错误
func (f *fakeProduct) reserve(ctx context.Context, tx *sql.Tx, req *event.StockReserveRequest) (event.StockOpResponse, error) {
	res, err := tx.ExecContext(ctx,
		"UPDATE stock SET available = available - ?, reserved = reserved + ? WHERE sku_id = ? AND available >= ? AND deleted_at IS NULL",
		req.Count, req.Count, req.SkuID, req.Count)
	if err != nil {
		return event.StockOpResponse{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return event.StockOpResponse{}, fmt.Errorf("商品 SKU %d 库存不足", req.SkuID)
	}
	return event.StockOpResponse{}, nil
}

// deduct 扣减库存（total-，reserved-），预留不足返回错误
func (f *fakeProduct) deduct(ctx context.Context, tx *sql.Tx, req *event.StockReserveRequest) (event.StockOpResponse, error) {
	res, err := tx.ExecContext(ctx,
		"UPDATE stock SET total = total - ?, reserved = reserved - ? WHERE sku_id = ? AND reserved >= ? AND deleted_at IS NULL",
		req.Count, req.Count, req.SkuID, req.Count)
	if err != nil {
		return event.StockOpResponse{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return event.StockOpResponse{}, fmt.Errorf("扣减库存失败 sku=%d: 预留库存不足", req.SkuID)
	}
	return event.StockOpResponse{}, nil
}

// release 释放预留（reserved-，available+）；条件不命中为幂等空操作（与 product 服务一致，不返回错误）
func (f *fakeProduct) release(ctx context.Context, tx *sql.Tx, req *event.StockReserveRequest) (event.StockOpResponse, error) {
	_, err := tx.ExecContext(ctx,
		"UPDATE stock SET reserved = reserved - ?, available = available + ? WHERE sku_id = ? AND reserved >= ? AND deleted_at IS NULL",
		req.Count, req.Count, req.SkuID, req.Count)
	return event.StockOpResponse{}, err
}

// add 逆扣减退还库存（total+，available+），库存记录不存在返回错误
func (f *fakeProduct) add(ctx context.Context, tx *sql.Tx, req *event.StockReserveRequest) (event.StockOpResponse, error) {
	res, err := tx.ExecContext(ctx,
		"UPDATE stock SET total = total + ?, available = available + ? WHERE sku_id = ? AND deleted_at IS NULL",
		req.Count, req.Count, req.SkuID)
	if err != nil {
		return event.StockOpResponse{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return event.StockOpResponse{}, fmt.Errorf("增加库存失败 sku=%d: 库存记录不存在", req.SkuID)
	}
	return event.StockOpResponse{}, nil
}

// get 查询库存三元组
func (f *fakeProduct) get(ctx context.Context, tx *sql.Tx, req *event.StockReserveRequest) (event.StockOpResponse, error) {
	var resp event.StockOpResponse
	err := tx.QueryRowContext(ctx,
		"SELECT total, available, reserved FROM stock WHERE sku_id = ? AND deleted_at IS NULL", req.SkuID,
	).Scan(&resp.Total, &resp.Available, &resp.Reserved)
	return resp, err
}

// TestTCC_CreatePay_Idempotent 验证 TCC 下单成功 -> 库存预留 -> 幂等重复下单不重复预留 -> 支付扣减 -> 重复支付幂等
func TestTCC_CreatePay_Idempotent(t *testing.T) {
	h := newHarness(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	skuID := uniqueSkuID()
	h.seedStock(t, skuID, 100)
	key := fmt.Sprintf("it-create-%d", time.Now().UnixNano())

	// 1. TCC 下单成功：订单待支付，库存预留 2 件（100/98/2）
	first := h.newOrderInput(key, skuID, 2)
	if err := h.uc.CreateTCC(ctx, first); err != nil {
		t.Fatalf("CreateTCC 首次下单失败: %v", err)
	}
	if first.ID == 0 {
		t.Fatal("下单成功但订单 ID 为 0")
	}
	created, err := h.uc.Get(ctx, first.ID)
	if err != nil {
		t.Fatalf("查询已创建订单失败: %v", err)
	}
	if created.Status != v1.OrderStatus_WaitPay {
		t.Fatalf("新建订单状态错误: 期望 WaitPay(%d), 实际 %d", v1.OrderStatus_WaitPay, created.Status)
	}
	if total, available, reserved := h.stockSnap(t, skuID); total != 100 || available != 98 || reserved != 2 {
		t.Fatalf("下单后库存错误: 期望 (100,98,2), 实际 (%d,%d,%d)", total, available, reserved)
	}

	// 2. 相同幂等键重复提交：直接返回首单，库存不发生第二次预留
	second := h.newOrderInput(key, skuID, 2)
	if err := h.uc.CreateTCC(ctx, second); err != nil {
		t.Fatalf("CreateTCC 幂等重复下单失败: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("幂等重复下单返回了不同订单: first=%d second=%d", first.ID, second.ID)
	}
	if total, available, reserved := h.stockSnap(t, skuID); total != 100 || available != 98 || reserved != 2 {
		t.Fatalf("重复下单后库存被错误变更: 期望 (100,98,2), 实际 (%d,%d,%d)", total, available, reserved)
	}

	// 3. 支付：扣减预留库存（total-、reserved-），订单流转已支付
	if err := h.uc.Pay(ctx, first.ID, h.userID); err != nil {
		t.Fatalf("支付订单失败: %v", err)
	}
	if total, available, reserved := h.stockSnap(t, skuID); total != 98 || available != 98 || reserved != 0 {
		t.Fatalf("支付后库存错误: 期望 (98,98,0), 实际 (%d,%d,%d)", total, available, reserved)
	}
	paid, err := h.uc.Get(ctx, first.ID)
	if err != nil {
		t.Fatalf("查询支付后订单失败: %v", err)
	}
	if paid.Status != v1.OrderStatus_Pay {
		t.Fatalf("支付后订单状态错误: 期望 Pay(%d), 实际 %d", v1.OrderStatus_Pay, paid.Status)
	}

	// 4. 重复支付幂等：库存不再变化
	if err := h.uc.Pay(ctx, first.ID, h.userID); err != nil {
		t.Fatalf("重复支付返回错误: %v", err)
	}
	if total, available, reserved := h.stockSnap(t, skuID); total != 98 || available != 98 || reserved != 0 {
		t.Fatalf("重复支付后库存被错误变更: 期望 (98,98,0), 实际 (%d,%d,%d)", total, available, reserved)
	}
}

// TestTCC_InsufficientStockRollback 验证库存不足时 TCC 全局回滚：不产生订单、库存无残留预留
func TestTCC_InsufficientStockRollback(t *testing.T) {
	h := newHarness(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	skuID := uniqueSkuID()
	h.seedStock(t, skuID, 1) // 仅有 1 件库存
	key := fmt.Sprintf("it-rollback-%d", time.Now().UnixNano())

	in := h.newOrderInput(key, skuID, 5) // 下单 5 件，Try 阶段必然失败
	err := h.uc.CreateTCC(ctx, in)
	if err == nil {
		t.Fatal("库存不足时 CreateTCC 应返回错误，实际返回 nil")
	}

	// 库存维持初始状态：预留 0、可用 1（Cancel 空补偿被屏障跳过，未产生错误释放）
	if total, available, reserved := h.stockSnap(t, skuID); total != 1 || available != 1 || reserved != 0 {
		t.Fatalf("回滚后库存错误: 期望 (1,1,0), 实际 (%d,%d,%d)", total, available, reserved)
	}
	// 未产生订单
	exist, err := h.repo.GetOrderByIdempotencyKey(ctx, h.userID, key)
	if err != nil {
		t.Fatalf("幂等键查询订单失败: %v", err)
	}
	if exist != nil {
		t.Fatalf("库存不足回滚后不应产生订单, 实际存在订单 ID=%d", exist.ID)
	}
	// DTM 已驱动 Cancel 补偿（屏障表存在 cancel 记录，空补偿被屏障裁决跳过）
	var cancelRows int
	if err := h.sqlDB.QueryRow(
		"SELECT COUNT(*) FROM dtm_barrier.barrier WHERE gid = ? AND op = 'cancel'",
		"tcc-"+in.OrderNo,
	).Scan(&cancelRows); err != nil {
		t.Fatalf("查询 DTM 屏障记录失败: %v", err)
	}
	if cancelRows < 1 {
		t.Fatalf("期望 DTM 执行 Cancel 补偿（屏障表应有 cancel 记录），实际 0 条")
	}
}

// TestTCC_CancelReleasesStock 验证待支付订单取消：TCC 预留库存被释放回可用量，订单流转取消终态
func TestTCC_CancelReleasesStock(t *testing.T) {
	h := newHarness(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	skuID := uniqueSkuID()
	h.seedStock(t, skuID, 100)
	key := fmt.Sprintf("it-cancel-%d", time.Now().UnixNano())

	in := h.newOrderInput(key, skuID, 2)
	if err := h.uc.CreateTCC(ctx, in); err != nil {
		t.Fatalf("CreateTCC 下单失败: %v", err)
	}
	if total, available, reserved := h.stockSnap(t, skuID); total != 100 || available != 98 || reserved != 2 {
		t.Fatalf("下单后库存错误: 期望 (100,98,2), 实际 (%d,%d,%d)", total, available, reserved)
	}

	// 用户取消订单：终态对账释放 TCC 预留（pay gid 无扣减记录，add 空补偿由屏障跳过）
	if err := h.uc.Cancel(ctx, in.ID, h.userID, "集成测试取消"); err != nil {
		t.Fatalf("取消订单失败: %v", err)
	}
	if total, available, reserved := h.stockSnap(t, skuID); total != 100 || available != 100 || reserved != 0 {
		t.Fatalf("取消后库存错误: 期望 (100,100,0), 实际 (%d,%d,%d)", total, available, reserved)
	}
	canceled, err := h.uc.Get(ctx, in.ID)
	if err != nil {
		t.Fatalf("查询取消后订单失败: %v", err)
	}
	if canceled.Status != v1.OrderStatus_Cancel {
		t.Fatalf("取消后订单状态错误: 期望 Cancel(%d), 实际 %d", v1.OrderStatus_Cancel, canceled.Status)
	}
}
