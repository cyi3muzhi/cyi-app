package data

import (
	"context"
	"cyi-app/app/service/order/internal/conf"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"cyi-app/app/service/order/internal/data/curd"
	mod "cyi-app/app/service/order/internal/data/models"

	productRpc "cyi-app/api/service/product/v1"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewOrderRepo,
	NewProductRepo,
	NewStockRepo,
	NewData,
	NewDB,
	NewRedisClient,
	NewProductServiceClient,
	NewStockServiceClient,
)

type Data struct {
	cache *redis.Client
	pc    productRpc.ProductClient
	sc    productRpc.StockClient
}

func NewData(c *conf.Data, db *gorm.DB, cache *redis.Client, logger log.Logger, pc productRpc.ProductClient, sc productRpc.StockClient) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	curd.SetDefault(db)
	return &Data{
		cache: cache,
		pc:    pc,
		sc:    sc,
	}, cleanup, nil
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
	if err := db.AutoMigrate(&mod.Order{}, &mod.OrderItem{}, &mod.OrderLogistics{}, &mod.PaymentLog{}); err != nil {
		panic(err)
	}

	// 仅在开发环境且文件不存在时执行生成
	if shouldGenerateCode() {
		generateCode(db)
	}

	return db
}

func shouldGenerateCode() bool {
	if os.Getenv("SERVICE_ENV") != "develop" {
		return false
	}
	return true
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
	g.Execute()
}

func NewRedisClient(c *conf.Data, logger log.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Network:      c.Redis.Network,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		Password:     c.Redis.Password,
	})
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelFunc()
	err := client.Ping(timeout).Err()
	if err != nil {
		panic(err)
	}
	return client
}
