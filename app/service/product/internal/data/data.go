package data

import (
	"cyi-app/app/service/product/internal/conf"
	"fmt"
	"os"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"cyi-app/app/service/product/internal/data/curd"
	mod "cyi-app/app/service/product/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewProductRepo,
	NewStockRepo,
	NewCategoryRepo,
	NewSpecGroupRepo,
	NewSkuRepo,
	NewData,
	NewDB,
)

type Data struct{}

func NewData(c *conf.Data, db *gorm.DB, logger log.Logger) (*Data, func(), error) {

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	curd.SetDefault(db)
	return &Data{}, cleanup, nil
}

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

	// 仅在非生产环境自动迁移表结构（生产环境禁止使用 AutoMigrate）
	env := os.Getenv("SERVICE_ENV")
	if env != "production" {
		if err := db.AutoMigrate(&mod.Product{}, &mod.Stock{}, &mod.Category{}, &mod.SpecGroup{}, &mod.SpecValue{}, &mod.Sku{}); err != nil {
			panic(err)
		}
	}

	// 仅在开发环境执行代码生成
	if env == "develop" {
		generateCode(db)
	}

	// 确保 FULLTEXT 索引存在（ngram 分词器，支持中文搜索）
	// 幂等执行：索引已存在时忽略 Duplicate key name 错误
	ensureFullTextIndexes(db)

	return db
}

// ensureFullTextIndexes 为 product 表创建 FULLTEXT 索引（ngram 分词器）
// 用于支持中文全文搜索，替代 LIKE 查询以提升性能
// 说明：MySQL FULLTEXT 的 MATCH() 列列表必须与索引列完全一致（不支持前缀匹配），
// 因此需要两个索引：
//   - ft_product_name    单列 (name)：供 SuggestProduct 的 MATCH(name) 使用
//   - ft_product_search  组合 (name, description)：供 ListProduct 的 MATCH(name, description) 使用
//
// 幂等：索引已存在时静默跳过
func ensureFullTextIndexes(db *gorm.DB) {
	indexes := []struct {
		name    string
		columns string
	}{
		{"ft_product_name", "`name`"},
		{"ft_product_search", "`name`,`description`"},
	}
	for _, idx := range indexes {
		sql := fmt.Sprintf("ALTER TABLE `product` ADD FULLTEXT INDEX `%s` (%s) WITH PARSER ngram", idx.name, idx.columns)
		if err := db.Exec(sql).Error; err != nil {
			// 索引已存在时静默跳过（MySQL 错误码 1061：Duplicate key name）
			if !isDuplicateIndexErr(err) {
				log.Warnf("create fulltext index %s failed: %v", idx.name, err)
			}
		}
	}
	// 清理旧版单列 description 索引（已被组合索引 ft_product_search 替代，避免冗余）
	dropLegacyDescIndex(db)
}

// dropLegacyDescIndex 删除旧版单列 description FULLTEXT 索引（ft_product_desc）
// 旧实现为 ListProduct 单独创建了 description 索引，但 MATCH(name, description)
// 无法命中单列索引（FULLTEXT 不支持前缀匹配），现已由组合索引替代
func dropLegacyDescIndex(db *gorm.DB) {
	_ = db.Exec("DROP INDEX `ft_product_desc` ON `product`").Error
}

// isDuplicateIndexErr 判断是否为"索引已存在"错误（幂等场景下静默跳过）
func isDuplicateIndexErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "Duplicate key name") || strings.Contains(msg, "1061")
}

func generateCode(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "../../internal/data/curd",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery,
	})
	g.UseDB(db)
	g.ApplyBasic(&mod.Product{}, &mod.Stock{}, &mod.Category{}, &mod.SpecGroup{}, &mod.SpecValue{}, &mod.Sku{})
	g.ApplyInterface(func(mod.IProduct) {}, mod.Product{})
	g.ApplyInterface(func(mod.IStock) {}, mod.Stock{})
	g.ApplyInterface(func(mod.ICategory) {}, mod.Category{})
	g.ApplyInterface(func(mod.ISpecGroup) {}, mod.SpecGroup{})
	g.ApplyInterface(func(mod.ISpecValue) {}, mod.SpecValue{})
	g.ApplyInterface(func(mod.ISku) {}, mod.Sku{})
	g.Execute()
}
