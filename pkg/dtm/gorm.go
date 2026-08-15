package dtm

import (
	"database/sql"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// GormFromTx 将 *sql.Tx 包装为绑定该事务的 *gorm.DB
// 供 DTM 屏障事务（barrier.CallWithDB 的 busiCall 内）执行 gen/curd 查询，
// 使业务写入与屏障记录处于同一数据库事务，保证幂等/回滚语义
// 包装失败返回 error（由 DTM 分支回调映射为 HTTP 500 触发重试），不 panic，避免进程崩溃
func GormFromTx(tx *sql.Tx) (*gorm.DB, error) {
	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      tx,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		return nil, err
	}
	return gdb, nil
}
