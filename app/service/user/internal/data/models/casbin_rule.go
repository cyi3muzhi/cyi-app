package models

import "time"

// CasbinRule Casbin 策略规则表
//
// 采用硬删除（不嵌入 DeletedAt）：策略删除即物理删除，
// 避免软删除记录长期占用唯一索引，导致同一策略无法重新授予。
// 唯一索引 uk_casbin_rule(ptype, v0, v1, v2) 用于保证重复部署、多实例并发初始化时
// 不会写入重复策略。
type CasbinRule struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Ptype     string `gorm:"column:ptype;type:varchar(100);not null;uniqueIndex:uk_casbin_rule,priority:1"`
	V0        string `gorm:"column:v0;type:varchar(100);default:'';uniqueIndex:uk_casbin_rule,priority:2"`
	V1        string `gorm:"column:v1;type:varchar(100);default:'';uniqueIndex:uk_casbin_rule,priority:3"`
	V2        string `gorm:"column:v2;type:varchar(100);default:'';uniqueIndex:uk_casbin_rule,priority:4"`
	V3        string `gorm:"column:v3;type:varchar(100);default:''"`
	V4        string `gorm:"column:v4;type:varchar(100);default:''"`
	V5        string `gorm:"column:v5;type:varchar(100);default:''"`
}

type ICasbinRule interface {
}
