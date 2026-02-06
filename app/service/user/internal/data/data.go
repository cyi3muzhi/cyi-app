package data

import (
	"context"
	"cyi-app/app/service/user/internal/conf"
	"cyi-app/pkg/util/rustfs"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"cyi-app/app/service/user/internal/data/curd"
	mod "cyi-app/app/service/user/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is dat  a providers.
var ProviderSet = wire.NewSet(
	NewUserRepo,
	NewData,
	NewDB,
	NewRedisClient,
	NewOssClient,
)

type Data struct {
	cache *redis.Client
	oss   *rustfs.RustFSClient
}

// NewData .
func NewData(c *conf.Data, db *gorm.DB, cache *redis.Client, oss *rustfs.RustFSClient, logger log.Logger) (*Data, func(), error) {

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	curd.SetDefault(db)
	return &Data{
		cache: cache,
		oss:   oss,
	}, cleanup, nil
}

func NewDB(c *conf.Data) *gorm.DB {
	// 连接数据库
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
	// 自动迁移：创建或更新数据库表结构
	if err := db.AutoMigrate(&mod.User{}); err != nil {
		panic(err)
	}
	// 初始化生成器（可选，用于生成查询代码）
	g := gen.NewGenerator(gen.Config{
		OutPath: "../../internal/data/curd",                // 生成文件的输出路径
		Mode:    gen.WithoutContext | gen.WithDefaultQuery, // 生成模式
	})
	// 添加数据库连接
	g.UseDB(db)
	// 生成模型的查询代码
	g.ApplyBasic(
		&mod.User{},
	)
	g.ApplyInterface(func(mod.IUser) {}, mod.User{})
	g.ApplyInterface(func(mod.IAddress) {}, mod.Address{})
	// 执行生成代码
	g.Execute()
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

func NewOssClient(cfg *conf.Data) (*rustfs.RustFSClient, error) {
	return rustfs.NewRustFSClient(&rustfs.Config{
		Endpoint:  cfg.Oss.Endpoint,
		AccessKey: cfg.Oss.AccessKey,
		SecretKey: cfg.Oss.SecretKey,
	})
}
