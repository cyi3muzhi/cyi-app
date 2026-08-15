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
func GormFromTx(tx *sql.Tx) *gorm.DB {
	gdb, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      tx,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		panic(err)
	}
	return gdb
}
