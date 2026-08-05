package data

import (
	"cyi-app/app/service/order/internal/conf"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"cyi-app/app/service/order/internal/data/curd"
	mod "cyi-app/app/service/order/internal/data/models"

	marketingRpc "cyi-app/api/service/marketing/v1"
	productRpc "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/hibiken/asynq"
	"github.com/nats-io/nats.go"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewOrderRepo,
	NewRefundRepo,
	NewProductRepo,
	NewStockRepo,
	NewData,
	NewDB,
	NewProductServiceClient,
	NewMarketingCouponServiceClient,
	NewNATSConn,
	NewAsynqClient,
	NewOrderTimeoutTaskRepo,
)

type Data struct {
	pc productRpc.ProductClient
	mc marketingRpc.CouponServiceClient
	nc *nats.Conn
	db *gorm.DB
	ac *asynq.Client // 投递延迟任务（订单超时自动取消）
}

func NewData(c *conf.Data, db *gorm.DB, logger log.Logger, pc productRpc.ProductClient, mc marketingRpc.CouponServiceClient, nc *nats.Conn, ac *asynq.Client) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if nc != nil {
			nc.Close()
		}
		if ac != nil {
			ac.Close()
		}
	}
	curd.SetDefault(db)
	return &Data{
		pc: pc,
		mc: mc,
		nc: nc,
		db: db,
		ac: ac,
	}, cleanup, nil
}

// NewNATSConn 创建 NATS 连接（单例）
// 优先从环境变量 NATS_URL 读取地址，默认使用 localhost（本地开发）
func NewNATSConn() *nats.Conn {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		panic(err)
	}
	return nc
}

func NewDB(c *conf.Data) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(err)
	}

	// 仅在非生产环境自动迁移表结构（生产环境禁止使用 AutoMigrate）
	env := os.Getenv("SERVICE_ENV")
	if env != "production" {
		if err := db.AutoMigrate(&mod.Order{}, &mod.OrderItem{}, &mod.OrderLogistics{}, &mod.PaymentLog{}, &mod.OrderRefund{}); err != nil {
			panic(err)
		}
		// 订单超时扫描复合索引：(status, create_time) 加速 task 服务 cron 兜底扫描
		// MySQL 不支持 CREATE INDEX IF NOT EXISTS，索引已存在时忽略报错
		if err := db.Exec("CREATE INDEX idx_orders_status_create_time ON `order`(status, create_time)").Error; err != nil {
			log.NewHelper(log.DefaultLogger).Debugf("create order status/create_time index skipped: %v", err)
		}
	}

	// 仅在开发环境执行代码生成
	if env == "develop" {
		generateCode(db)
	}

	return db
}

func generateCode(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "../../internal/data/curd",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery,
	})
	g.UseDB(db)
	g.ApplyBasic(
		&mod.Order{},
		&mod.OrderItem{},
		&mod.OrderLogistics{},
		&mod.PaymentLog{},
		&mod.OrderRefund{},
	)
	g.ApplyInterface(func(mod.IOrder) {},
		mod.Order{},
	)
	g.ApplyInterface(func(mod.IOrderItem) {},
		mod.OrderItem{},
	)
	g.ApplyInterface(func(mod.IOrderLogistics) {},
		mod.OrderLogistics{},
	)
	g.ApplyInterface(func(mod.IPaymentLog) {},
		mod.PaymentLog{},
	)
	g.ApplyInterface(func(mod.IOrderRefund) {},
		mod.OrderRefund{},
	)
	g.Execute()
}
