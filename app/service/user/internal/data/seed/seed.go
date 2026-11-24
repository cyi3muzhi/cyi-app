// Package seed 提供系统初始化种子数据。
//
// 种子数据在 user 服务启动时自动执行，用于保证一套全新的数据库开箱即可登录：
// 预置超级管理员账号、常用业务角色及其权限。
// 所有写入操作均为幂等设计——记录已存在时不会重复创建，也不会覆盖后台已修改的数据。
package seed

import (
	"fmt"

	"cyi-app/app/service/user/internal/data/models"
	"cyi-app/app/service/user/internal/pkg/encryption"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// DefaultAdminUsername 默认超级管理员用户名
	DefaultAdminUsername = "admin"
	// DefaultAdminPassword 默认超级管理员密码（首次初始化使用，部署后请立即修改）
	DefaultAdminPassword = "admin123"
	// SuperAdminSlug 超级管理员角色标识
	SuperAdminSlug = "super_admin"
)

// roleSeed 角色种子定义
type roleSeed struct {
	Slug        string              // 角色标识，同时作为 Casbin 主体（sub）
	Name        string              // 角色显示名称
	Desc        string              // 角色描述
	Color       string              // 角色标签颜色
	All         bool                // 是否拥有全部权限（仅超级管理员）
	Permissions map[string][]string // 权限清单：模块 key -> 动作 key 列表
}

// roles 预置角色清单。
//
// Permissions 中的模块 key 与动作 key 必须与 data 包 permModules 中的定义保持一致；
// 写入 Casbin 策略时 obj 拼为 "/<module>/<action>"，act 为 <action>，
// 与后台「角色权限」保存逻辑保持同一套规则。
var roles = []roleSeed{
	{
		Slug:  SuperAdminSlug,
		Name:  "超级管理员",
		Desc:  "拥有所有系统权限",
		Color: "#F56C6C",
		All:   true,
	},
	{
		Slug:  "operator",
		Name:  "运营",
		Desc:  "负责商品运营、活动配置与消息推送",
		Color: "#409EFF",
		Permissions: map[string][]string{
			"data":     {"dashboard", "report", "export"},
			"products": {"view", "publish"},
			"orders":   {"view"},
			"coupons":  {"view", "create", "edit", "void"},
			"banners":  {"view", "add", "edit", "delete"},
			"messages": {"view", "send", "delete"},
			"chat":     {"view"},
			"users":    {"view"},
		},
	},
	{
		Slug:  "product_admin",
		Name:  "商品管理员",
		Desc:  "负责商品与分类的维护",
		Color: "#67C23A",
		Permissions: map[string][]string{
			"products":   {"view", "add", "edit", "delete", "publish"},
			"categories": {"view", "add", "edit", "delete"},
			"data":       {"dashboard"},
		},
	},
	{
		Slug:  "order_admin",
		Name:  "订单管理员",
		Desc:  "负责订单处理、发货与退款",
		Color: "#E6A23C",
		Permissions: map[string][]string{
			"orders": {"view", "edit", "refund", "ship", "cancel"},
			"data":   {"dashboard", "report"},
			"users":  {"view"},
		},
	},
	{
		Slug:  "customer_service",
		Name:  "客服",
		Desc:  "负责客服工作台会话接待",
		Color: "#909399",
		Permissions: map[string][]string{
			"chat":     {"view", "accept", "transfer", "close"},
			"users":    {"view"},
			"messages": {"view", "send"},
		},
	},
}

// Run 执行种子数据初始化。
//
// 由 data.NewDB 在服务启动时调用，所有步骤幂等：
//  1. 初始化角色元数据（role_meta）
//  2. 初始化角色权限策略（casbin_rule，p 策略）
//  3. 初始化默认超级管理员账号（admin_user）
//  4. 同步所有管理员账号的角色绑定（casbin_rule，g 策略）
//
// 数据表尚未创建时（例如生产环境未执行迁移）自动跳过，不影响服务启动。
func Run(db *gorm.DB, logger log.Logger) error {
	helper := log.NewHelper(logger)

	if !db.Migrator().HasTable(&models.RoleMeta{}) ||
		!db.Migrator().HasTable(&models.AdminUser{}) ||
		!db.Migrator().HasTable(&models.CasbinRule{}) {
		helper.Warn("seed: 数据表尚未创建，跳过种子数据初始化")
		return nil
	}

	// 兼容旧版本数据：清理历史软删除策略、删除重复策略，保证唯一索引可用
	if err := Prepare(db); err != nil {
		return err
	}

	if err := seedRoles(db, helper); err != nil {
		return err
	}
	if err := seedAdminUser(db, helper); err != nil {
		return err
	}
	if err := syncAdminRoles(db, helper); err != nil {
		return err
	}

	helper.Info("seed: 种子数据初始化完成")
	return nil
}

// Prepare 在自动迁移前清理历史遗留数据，确保 Casbin 策略唯一索引能够成功创建。
//
// 旧版本 CasbinRule 使用软删除，历史软删除记录会占用唯一索引，
// 且可能已存在重复策略。这里做两件事（均幂等）：
//  1. 物理删除历史软删除的策略行；
//  2. 删除重复策略，同一 (ptype, v0, v1, v2) 仅保留 ID 最小的一条。
//
// 数据表不存在时直接返回。生产环境启动时也会执行。
func Prepare(db *gorm.DB) error {
	if !db.Migrator().HasTable(&models.CasbinRule{}) {
		return nil
	}
	// 旧表才有 deleted_at 列，新表结构已移除该列
	if db.Migrator().HasColumn(&models.CasbinRule{}, "deleted_at") {
		if err := db.Exec("DELETE FROM casbin_rule WHERE deleted_at IS NOT NULL").Error; err != nil {
			return fmt.Errorf("清理历史软删除策略失败: %w", err)
		}
	}
	// MySQL 自连接删除重复行，保留 id 最小的一条
	if err := db.Exec(`
		DELETE c1 FROM casbin_rule c1
		INNER JOIN casbin_rule c2
			ON c1.ptype = c2.ptype
		   AND c1.v0 = c2.v0
		   AND c1.v1 = c2.v1
		   AND c1.v2 = c2.v2
		   AND c1.id > c2.id`).Error; err != nil {
		return fmt.Errorf("清理重复策略失败: %w", err)
	}
	return nil
}

// seedRoles 初始化角色元数据与 Casbin 权限策略
func seedRoles(db *gorm.DB, helper *log.Helper) error {
	for _, r := range roles {
		// Unscoped 查询：需要区分「从未创建」和「已被后台删除」两种情况
		var meta models.RoleMeta
		err := db.Unscoped().Where("slug = ?", r.Slug).First(&meta).Error
		switch {
		case err == gorm.ErrRecordNotFound:
			// 首次部署：创建角色元数据（slug 唯一索引，并发冲突忽略）
			meta = models.RoleMeta{Slug: r.Slug, Name: r.Name, Desc: r.Desc, Color: r.Color}
			result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&meta)
			if result.Error != nil {
				return fmt.Errorf("创建角色 %s 失败: %w", r.Slug, result.Error)
			}
			if result.RowsAffected > 0 {
				helper.Infof("seed: 已创建角色 %s (%s)", r.Name, r.Slug)
			}
		case err != nil:
			return fmt.Errorf("查询角色 %s 失败: %w", r.Slug, err)
		case meta.DeletedAt.Valid:
			// 角色已被后台删除：尊重删除操作，跳过权限初始化，避免重启后复活
			continue
		}

		// 权限策略：超级管理员使用通配符，其余角色按清单逐项授权
		if r.All {
			if err := ensurePolicy(db, r.Slug, "*", "*"); err != nil {
				return fmt.Errorf("初始化角色 %s 权限失败: %w", r.Slug, err)
			}
			continue
		}
		for module, actions := range r.Permissions {
			for _, action := range actions {
				obj := "/" + module + "/" + action
				if err := ensurePolicy(db, r.Slug, obj, action); err != nil {
					return fmt.Errorf("初始化角色 %s 权限 %s 失败: %w", r.Slug, obj, err)
				}
			}
		}
	}
	return nil
}

// seedAdminUser 初始化默认超级管理员账号，并绑定 super_admin 角色
func seedAdminUser(db *gorm.DB, helper *log.Helper) error {
	var admin models.AdminUser
	// Unscoped 查询：被软删除的默认管理员也需要识别并恢复
	err := db.Unscoped().Where("username = ?", DefaultAdminUsername).First(&admin).Error
	switch {
	case err == gorm.ErrRecordNotFound:
		hashedPwd, hashErr := encryption.EncryptPassword(DefaultAdminPassword)
		if hashErr != nil {
			return fmt.Errorf("默认管理员密码加密失败: %w", hashErr)
		}
		candidate := models.AdminUser{
			Username: DefaultAdminUsername,
			Password: hashedPwd,
			Name:     "管理员",
			Role:     SuperAdminSlug,
			Email:    "admin@yuexuan.com",
			Status:   1,
		}
		// 依赖 username 唯一索引，多实例并发初始化时冲突忽略
		result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&candidate)
		if result.Error != nil {
			return fmt.Errorf("创建默认管理员失败: %w", result.Error)
		}
		if result.RowsAffected > 0 {
			admin = candidate
			helper.Infof("seed: 已创建默认管理员账号 %s/%s（请尽快修改密码）", DefaultAdminUsername, DefaultAdminPassword)
		} else {
			// 已被其他实例抢先创建，重新查询获取真实用户 ID
			if err := db.Where("username = ?", DefaultAdminUsername).First(&admin).Error; err != nil {
				return fmt.Errorf("查询默认管理员失败: %w", err)
			}
		}
	case err != nil:
		return fmt.Errorf("查询默认管理员失败: %w", err)
	case admin.DeletedAt.Valid:
		// 默认管理员曾被软删除，恢复账号并确保其仍为超级管理员
		if err := db.Unscoped().Model(&models.AdminUser{}).
			Where("id = ?", admin.ID).
			Updates(map[string]interface{}{"deleted_at": nil, "role": SuperAdminSlug}).Error; err != nil {
			return fmt.Errorf("恢复默认管理员失败: %w", err)
		}
		helper.Infof("seed: 已恢复被删除的默认管理员账号 %s", DefaultAdminUsername)
	}

	// 分组策略使用真实用户 ID，避免旧实现硬编码 ID=1 导致的错位
	return ensureGrouping(db, fmt.Sprintf("%d", admin.ID), SuperAdminSlug)
}

// syncAdminRoles 将所有管理员的角色同步到 Casbin 分组策略。
// 用于兼容历史上只写了 admin_user.role、尚未写入 g 策略的账号。
func syncAdminRoles(db *gorm.DB, helper *log.Helper) error {
	var admins []models.AdminUser
	if err := db.Find(&admins).Error; err != nil {
		return fmt.Errorf("查询管理员列表失败: %w", err)
	}
	for _, admin := range admins {
		if admin.Role == "" {
			continue
		}
		if err := ensureGrouping(db, fmt.Sprintf("%d", admin.ID), admin.Role); err != nil {
			helper.Warnf("seed: 为管理员 %d 绑定角色 %s 失败: %v", admin.ID, admin.Role, err)
		}
	}
	return nil
}

// ensurePolicy 确保一条 Casbin p 策略存在（幂等）
func ensurePolicy(db *gorm.DB, sub, obj, act string) error {
	var count int64
	if err := db.Model(&models.CasbinRule{}).
		Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", sub, obj, act).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	// 依赖唯一索引 uk_casbin_rule，多实例并发初始化时冲突忽略
	return db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.CasbinRule{Ptype: "p", V0: sub, V1: obj, V2: act}).Error
}

// ensureGrouping 确保一条 Casbin g 分组策略存在（幂等）
func ensureGrouping(db *gorm.DB, userID, role string) error {
	var count int64
	if err := db.Model(&models.CasbinRule{}).
		Where("ptype = ? AND v0 = ? AND v1 = ?", "g", userID, role).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	// 依赖唯一索引 uk_casbin_rule，多实例并发初始化时冲突忽略
	return db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.CasbinRule{Ptype: "g", V0: userID, V1: role}).Error
}
