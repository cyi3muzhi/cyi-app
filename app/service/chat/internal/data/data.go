package data

import (
	"cyi-app/app/service/chat/internal/conf"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"cyi-app/app/service/chat/internal/data/curd"
	mod "cyi-app/app/service/chat/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
)

var ProviderSet = wire.NewSet(
	NewConversationRepo,
	NewMessageRepo,
	NewChatMessagePublisher,
	NewConversationEventPublisher,
	NewBladesClient,
	NewUserMessageClient,
	NewData,
	NewDB,
	NewNATSConn,
	NewRedisClient,
)

type Data struct {
	nc  *nats.Conn
	rdb redis.UniversalClient
}

// NewData 初始化数据层（M3：持有 NATS 连接与 Redis，用于事件发布与在线状态查询）
func NewData(c *conf.Data, db *gorm.DB, nc *nats.Conn, rdb redis.UniversalClient, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		if nc != nil {
			nc.Close()
		}
		if rdb != nil {
			_ = rdb.Close()
		}
		log.NewHelper(logger).Info("closing the data resources")
	}
	curd.SetDefault(db)
	return &Data{nc: nc, rdb: rdb}, cleanup, nil
}

// NewRedisClient 创建 Redis 客户端（在线状态/未读数，M3 起使用）
func NewRedisClient(c *conf.Data) redis.UniversalClient {
	return redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Network:      c.Redis.Network,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		Password:     c.Redis.Password,
	})
}

// NewDB 创建数据库连接（非生产环境自动迁移表结构）
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

	env := os.Getenv("SERVICE_ENV")
	if env != "production" {
		if err := db.AutoMigrate(&mod.Conversation{}, &mod.ChatMessage{}); err != nil {
			panic(err)
		}
	}

	if env == "develop" {
		generateCode(db)
	}

	return db
}

// generateCode 使用 gorm gen 生成查询代码（仅 develop 环境执行）
func generateCode(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "../../internal/data/curd",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery,
	})
	g.UseDB(db)
	g.ApplyBasic(
		&mod.Conversation{},
		&mod.ChatMessage{},
	)
	g.ApplyInterface(func(mod.IConversation) {},
		mod.Conversation{},
	)
	g.ApplyInterface(func(mod.IChatMessage) {},
		mod.ChatMessage{},
	)
	g.Execute()
}
