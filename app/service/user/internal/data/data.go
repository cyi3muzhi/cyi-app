package data

import (
	"context"
	"cyi-app/app/service/user/internal/conf"
	"cyi-app/app/service/user/internal/data/seed"
	"cyi-app/pkg/util/rustfs"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"cyi-app/app/service/user/internal/data/curd"
	mod "cyi-app/app/service/user/internal/data/models"

	"github.com/casbin/casbin/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewUserRepo,
	NewAddressRepo,
	NewMessageRepo,
	NewAdminUserRepo,
	NewRoleRepo,
	NewSettingsRepo,
	NewBannerRepo,
	NewOperationLogRepo,
	NewOperationLogConsumer,
	NewNATSConn,
	NewData,
	NewDB,
	NewRedisClient,
	NewOssClient,
	NewCasbinEnforcer,
)

type Data struct {
	db       *gorm.DB
	cache    *redis.Client
	oss      *rustfs.RustFSClient
	enforcer *casbin.Enforcer
	nc       *nats.Conn
}

// NewData .
func NewData(c *conf.Data, db *gorm.DB, cache *redis.Client, oss *rustfs.RustFSClient, enforcer *casbin.Enforcer, nc *nats.Conn, logger log.Logger) (*Data, func(), error) {

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
		if nc != nil {
			nc.Close()
		}
	}
	curd.SetDefault(db)
	return &Data{
		db:       db,
		cache:    cache,
		oss:      oss,
		enforcer: enforcer,
		nc:       nc,
	}, cleanup, nil
}

// NewNATSConn 创建 NATS 连接（单例）
// 优先从环境变量 NATS_URL 读取地址（Docker 环境应配置为 nats://nats:4222），默认使用 localhost（本地开发）
// 启动时基础设施容器可能尚未就绪，最多重试 5 次；仍失败则返回 nil 降级（不 panic，避免拖垮整个服务）
func NewNATSConn() *nats.Conn {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}
	for i := 0; i < 5; i++ {
		nc, err := nats.Connect(natsURL)
		if err == nil {
			return nc
		}
		time.Sleep(2 * time.Second)
	}
	// 连接失败不 panic：操作日志消费降级跳过（OperationLogConsumer.Start 内部有 nil 保护）
	return nil
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

	// 仅在非生产环境自动迁移表结构（生产环境禁止使用 AutoMigrate）
	env := os.Getenv("SERVICE_ENV")
	if env != "production" {
		// 迁移前清理历史软删除与重复策略，保证 Casbin 唯一索引能够成功创建
		if err := seed.Prepare(db); err != nil {
			log.NewHelper(log.DefaultLogger).Warnf("清理历史策略数据失败: %v", err)
		}
		if err := db.AutoMigrate(&mod.User{}, &mod.Address{}, &mod.AdminUser{}, &mod.CasbinRule{}, &mod.RoleMeta{}, &mod.SettingsRow{}, &mod.BannersRow{}, &mod.AdminOperationLog{}, &mod.Message{}, &mod.UserMessage{}); err != nil {
			panic(err)
		}
	}

	// 初始化种子数据（角色、默认管理员），幂等执行，生产环境同样生效，
	// 保证一套全新的数据库部署后可以直接使用默认账号登录。
	if err := seed.Run(db, log.DefaultLogger); err != nil {
		log.NewHelper(log.DefaultLogger).Warnf("初始化种子数据失败: %v", err)
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
	g.ApplyBasic(&mod.User{}, &mod.AdminUser{})
	g.ApplyInterface(func(mod.IUser) {}, mod.User{})
	g.ApplyInterface(func(mod.IAdminUser) {}, mod.AdminUser{})
	g.ApplyInterface(func(mod.IRoleMeta) {}, mod.RoleMeta{})
	g.ApplyInterface(func(mod.ICasbinRule) {}, mod.CasbinRule{})
	g.Execute()
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
