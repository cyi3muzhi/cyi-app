package data

import (
	"context"
	"cyi-app/app/service/user/internal/conf"
	mod "cyi-app/app/service/user/internal/data/models"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewUserRepo,
	NewData,
	NewDB,
	NewRedisClient,
)

type Data struct {
	db    func(ctx context.Context) *gorm.DB
	cache *redis.Client
}

// NewData .
func NewData(c *conf.Data, db *gorm.DB, cache *redis.Client, logger log.Logger) (*Data, func(), error) {

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	// 获取事务数据库
	getTxDB := func(ctx context.Context) *gorm.DB {
		fmt.Println("ctx:", ctx.Value("tx_db"))
		if txDB, ok := ctx.Value("tx_db").(*gorm.DB); ok {
			fmt.Println("txDB:", txDB)
			return txDB
		}
		return db
	}

	return &Data{
		db:    getTxDB,
		cache: cache,
	}, cleanup, nil
}

func NewDB(c *conf.Data) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			// 表名不使用复数
			SingularTable: true,
		},
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&mod.User{}); err != nil {
		panic(err)
	}
	return db
}

// NewRedisClient .
func NewRedisClient(c *conf.Data, logger log.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Network:      c.Redis.Network,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		Password:     c.Redis.Password,
	})
	// client.AddHook(redisotel.NewTracingHook())
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*2)
	defer cancelFunc()
	err := client.Ping(timeout).Err()
	if err != nil {
		panic(err)
	}
	return client
}
