package data

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupProductDataTestDB 初始化内存 SQLite 并迁移商品相关模型。
// 使用纯 Go 实现的 sqlite 驱动，避免对 CGO 的依赖；
// 通过限制连接池只有一个连接，确保内存数据库在 curd 与直接 db 操作之间保持一致可见性；
// 每个测试使用独立的数据库名称，保证隔离性。
func setupProductDataTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	sqlDB, err := sql.Open(sqlite.DriverName, dsn)
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	// 强制只使用一个连接，避免 SQLite 内存数据库在不同连接间出现数据不可见的问题
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	db, err := gorm.Open(sqlite.Dialector{Conn: sqlDB}, &gorm.Config{})
	if err != nil {
		t.Fatalf("open gorm failed: %v", err)
	}
	err = db.AutoMigrate(&mod.Product{}, &mod.Category{}, &mod.SpecGroup{}, &mod.SpecValue{}, &mod.Sku{})
	if err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}
	curd.SetDefault(db)
	return db
}

// requireNoError 辅助断言：非预期错误时立即终止当前测试
func requireNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
