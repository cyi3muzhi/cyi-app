package data

import (
	"context"
	"cyi-app/app/service/user/internal/conf"
	"cyi-app/app/service/user/internal/pkg/encryption"
	"cyi-app/pkg/util/rustfs"
	"os"
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
	NewAdminUserRepo,
	NewData,
	NewDB,
	NewRedisClient,
	NewOssClient,
)

type Data struct {
	db    *gorm.DB
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
		db:    db,
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
	if err := db.AutoMigrate(&mod.User{}, &mod.AdminUser{}); err != nil {
		panic(err)
	}

	// 初始化默认管理员（仅表为空时插入）
	initDefaultAdmin(db)

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
	g.ApplyBasic(&mod.User{}, &mod.AdminUser{})
	g.ApplyInterface(func(mod.IUser) {}, mod.User{})
	g.ApplyInterface(func(mod.IAdminUser) {}, mod.AdminUser{})
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

// initDefaultAdmin 初始化默认管理员账号（仅首次启动时插入）
func initDefaultAdmin(db *gorm.DB) {
	var count int64
	db.Model(&mod.AdminUser{}).Count(&count)
	if count > 0 {
		return
	}
	hashedPwd, err := encryption.EncryptPassword("admin123")
	if err != nil {
		log.NewHelper(log.DefaultLogger).Warnf("默认管理员密码加密失败: %v", err)
		return
	}
	now := time.Now().Unix()
	admin := &mod.AdminUser{
		Username:  "admin",
		Password:  hashedPwd,
		Name:      "管理员",
		Role:      "超级管理员",
		Email:     "admin@yuexuan.com",
		Status:    1,
		LastLogin: 0,
	}
	admin.CreateTime = now
	admin.UpdateTime = now
	if err := db.Create(admin).Error; err != nil {
		log.NewHelper(log.DefaultLogger).Warnf("初始化默认管理员失败: %v", err)
	} else {
		log.NewHelper(log.DefaultLogger).Infof("已创建默认管理员账号: admin/admin123")
	}
}
