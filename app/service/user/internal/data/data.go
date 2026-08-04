package data

import (
	"context"
	"cyi-app/app/service/user/internal/conf"
	"cyi-app/app/service/user/internal/pkg/encryption"
	"cyi-app/pkg/util/rustfs"
	"fmt"
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
	"github.com/redis/go-redis/v9"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewUserRepo,
	NewAddressRepo,
	NewAdminUserRepo,
	NewRoleRepo,
	NewSettingsRepo,
	NewBannerRepo,
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
}

// NewData .
func NewData(c *conf.Data, db *gorm.DB, cache *redis.Client, oss *rustfs.RustFSClient, enforcer *casbin.Enforcer, logger log.Logger) (*Data, func(), error) {

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	curd.SetDefault(db)
	return &Data{
		db:       db,
		cache:    cache,
		oss:      oss,
		enforcer: enforcer,
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

	// 仅在非生产环境自动迁移表结构（生产环境禁止使用 AutoMigrate）
	env := os.Getenv("SERVICE_ENV")
	if env != "production" {
		if err := db.AutoMigrate(&mod.User{}, &mod.Address{}, &mod.AdminUser{}, &mod.CasbinRule{}, &mod.RoleMeta{}, &mod.SettingsRow{}, &mod.BannersRow{}); err != nil {
			panic(err)
		}

		// 初始化默认管理员（仅表为空时插入）
		initDefaultAdmin(db)

		// 初始化 Casbin 默认策略（仅表为空时插入）
		initCasbinPolicies(db)

		// 同步所有管理员的角色到 Casbin 分组策略
		syncAdminUserRoles(db)
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
		Role:      "super_admin",
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

// initCasbinPolicies 初始化 Casbin 默认策略（仅首次启动时 casbin_rule 表为空时插入）
func initCasbinPolicies(db *gorm.DB) {
	var count int64
	db.Model(&mod.CasbinRule{}).Count(&count)
	if count > 0 {
		return
	}

	rules := []mod.CasbinRule{
		// 超级管理员拥有所有权限（通配符）
		{Ptype: "p", V0: "super_admin", V1: "*", V2: "*"},
		// 将默认管理员用户（ID=1）分配给超级管理员角色
		{Ptype: "g", V0: "1", V1: "super_admin"},
	}
	if err := db.Create(&rules).Error; err != nil {
		log.NewHelper(log.DefaultLogger).Warnf("初始化 Casbin 策略失败: %v", err)
	} else {
		log.NewHelper(log.DefaultLogger).Infof("已初始化 Casbin 默认策略")
	}

	// 同时写入超级管理员的角色元数据
	var metaCount int64
	db.Model(&mod.RoleMeta{}).Count(&metaCount)
	if metaCount == 0 {
		db.Create(&mod.RoleMeta{Slug: "super_admin", Name: "超级管理员", Desc: "拥有所有系统权限"})
	}
}

// syncAdminUserRoles 同步所有管理员的角色到 Casbin 分组策略
// 确保已有管理员的角色也能在 Casbin 中被正确统计
func syncAdminUserRoles(db *gorm.DB) {
	var admins []mod.AdminUser
	if err := db.Find(&admins).Error; err != nil {
		log.NewHelper(log.DefaultLogger).Warnf("查询管理员列表失败: %v", err)
		return
	}
	for _, admin := range admins {
		if admin.Role == "" {
			continue
		}
		// 检查是否已有分组策略
		var count int64
		db.Model(&mod.CasbinRule{}).
			Where("ptype = ? AND v0 = ? AND v1 = ?", "g", fmt.Sprintf("%d", admin.ID), admin.Role).
			Count(&count)
		if count == 0 {
			if err := db.Create(&mod.CasbinRule{Ptype: "g", V0: fmt.Sprintf("%d", admin.ID), V1: admin.Role}).Error; err != nil {
				log.NewHelper(log.DefaultLogger).Warnf("为用户 %d 添加 Casbin 分组策略失败: %v", admin.ID, err)
			} else {
				log.NewHelper(log.DefaultLogger).Infof("已为用户 %d 添加 Casbin 角色: %s", admin.ID, admin.Role)
			}
		}
	}
}
