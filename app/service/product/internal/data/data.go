package data

import (
	"context"
	"cyi-app/app/service/product/internal/conf"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewProductRepo,
	NewStockRepo,
	NewData,
	NewDB,
	NewRedisClient,
)

type Data struct {
	cache *redis.Client
}

func NewData(c *conf.Data, db *gorm.DB, cache *redis.Client, logger log.Logger) (*Data, func(), error) {

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	curd.SetDefault(db)
	return &Data{
		cache: cache,
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
	if err := db.AutoMigrate(&mod.Product{}, &mod.Stock{}); err != nil {
		panic(err)
	}

	// 仅在开发环境且文件不存在时执行生成
	if shouldGenerateCode() {
		generateCode(db)
	}

	return db
}

func shouldGenerateCode() bool {
	// 检查环境变量
	if os.Getenv("SERVICE_ENV") != "develop" {
		return false
	}

	// 检查生成目录是否已存在内容
	// _, err := os.Stat("../../internal/data/curd/user.gen.go")
	return true
}

func generateCode(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "../../internal/data/curd",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery,
	})
	g.UseDB(db)
	g.ApplyBasic(&mod.Product{}, &mod.Stock{})
	g.ApplyInterface(func(mod.IProduct) {}, mod.Product{})
	g.ApplyInterface(func(mod.IStock) {}, mod.Stock{})
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
