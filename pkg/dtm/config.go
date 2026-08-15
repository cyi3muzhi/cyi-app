package dtm

import (
	"database/sql"

	"gorm.io/gorm"
)

// ServerAddr DTM 服务地址（命名类型，供 wire 区分多个 string 依赖）
type ServerAddr string

// BusiURL DTM 回调本服务分支的地址（Confirm/Cancel 由 DTM 服务回调，命名类型供 wire 区分）
type BusiURL string

// LocalURL 本服务进程访问自己分支的地址（Try 由本服务进程内发起，命名类型供 wire 区分）
type LocalURL string

// BarrierDB 返回 DTM 屏障表所在数据库连接（*sql.DB，供 dtmcli.BranchBarrier 使用）
// 屏障表默认位于 dtm_barrier 库（见 deploy/docker-compose/mysql/init/01-dtm.sql）
func BarrierDB(db *gorm.DB) *sql.DB {
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	return sqlDB
}
