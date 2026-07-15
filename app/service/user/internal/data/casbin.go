package data

import (
	"fmt"

	"cyi-app/app/service/user/internal/data/models"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"gorm.io/gorm"
)

// gormAdapter 自定义 GORM 适配器（实现 casbin/v2 的 persist.Adapter 接口）
type gormAdapter struct {
	db *gorm.DB
}

// newGormAdapter 创建 GORM 适配器实例（表迁移由 data.NewDB 统一处理）
func newGormAdapter(db *gorm.DB) *gormAdapter {
	return &gormAdapter{db: db}
}

// LoadPolicy 从数据库加载所有策略到 Casbin 模型中
func (a *gormAdapter) LoadPolicy(model model.Model) error {
	var rules []models.CasbinRule
	if err := a.db.Find(&rules).Error; err != nil {
		return err
	}
	for _, rule := range rules {
		line := rule.Ptype
		if rule.V0 != "" {
			line += ", " + rule.V0
		}
		if rule.V1 != "" {
			line += ", " + rule.V1
		}
		if rule.V2 != "" {
			line += ", " + rule.V2
		}
		if rule.V3 != "" {
			line += ", " + rule.V3
		}
		if rule.V4 != "" {
			line += ", " + rule.V4
		}
		if rule.V5 != "" {
			line += ", " + rule.V5
		}
		persist.LoadPolicyLine(line, model)
	}
	return nil
}

// SavePolicy 将 Casbin 模型中的所有策略保存到数据库
func (a *gormAdapter) SavePolicy(model model.Model) error {
	a.db.Where("1 = 1").Delete(&models.CasbinRule{})

	var rules []models.CasbinRule
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			rules = append(rules, a.ruleFromLine(ptype, rule))
		}
	}
	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			rules = append(rules, a.ruleFromLine(ptype, rule))
		}
	}
	if len(rules) > 0 {
		return a.db.Create(&rules).Error
	}
	return nil
}

// AddPolicy 添加单条策略到数据库
func (a *gormAdapter) AddPolicy(sec string, ptype string, rule []string) error {
	r := a.ruleFromLine(ptype, rule)
	return a.db.Create(&r).Error
}

// RemovePolicy 从数据库删除单条策略
func (a *gormAdapter) RemovePolicy(sec string, ptype string, rule []string) error {
	return a.deleteRule(ptype, rule).Error
}

// RemoveFilteredPolicy 按条件过滤删除策略
func (a *gormAdapter) RemoveFilteredPolicy(sec string, ptype string, fieldIndex int, fieldValues ...string) error {
	query := a.db.Where("ptype = ?", ptype)
	for i, val := range fieldValues {
		if val != "" {
			col := a.getColumnName(fieldIndex + i)
			query = query.Where(col+" = ?", val)
		}
	}
	return query.Delete(&models.CasbinRule{}).Error
}

// ruleFromLine 将策略行转换为 CasbinRule 结构体
func (a *gormAdapter) ruleFromLine(ptype string, rule []string) models.CasbinRule {
	r := models.CasbinRule{Ptype: ptype}
	if len(rule) > 0 {
		r.V0 = rule[0]
	}
	if len(rule) > 1 {
		r.V1 = rule[1]
	}
	if len(rule) > 2 {
		r.V2 = rule[2]
	}
	if len(rule) > 3 {
		r.V3 = rule[3]
	}
	if len(rule) > 4 {
		r.V4 = rule[4]
	}
	if len(rule) > 5 {
		r.V5 = rule[5]
	}
	return r
}

// deleteRule 按条件删除策略
func (a *gormAdapter) deleteRule(ptype string, rule []string) *gorm.DB {
	query := a.db.Where("ptype = ?", ptype)
	for i, val := range rule {
		if val != "" {
			col := a.getColumnName(i)
			query = query.Where(col+" = ?", val)
		}
	}
	return query.Delete(&models.CasbinRule{})
}

// getColumnName 根据索引获取对应的数据库列名
func (a *gormAdapter) getColumnName(i int) string {
	cols := []string{"v0", "v1", "v2", "v3", "v4", "v5"}
	if i >= 0 && i < len(cols) {
		return cols[i]
	}
	return "v0"
}

// NewCasbinEnforcer 创建 Casbin Enforcer 实例（以 GORM + MySQL 替代文件存储）
func NewCasbinEnforcer(db *gorm.DB) (*casbin.Enforcer, error) {
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && (r.act == p.act || p.act == "*")
`
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("加载 Casbin 模型失败: %w", err)
	}

	a := newGormAdapter(db)
	e, err := casbin.NewEnforcer(m, a)
	if err != nil {
		return nil, fmt.Errorf("创建 Casbin Enforcer 失败: %w", err)
	}
	return e, nil
}
