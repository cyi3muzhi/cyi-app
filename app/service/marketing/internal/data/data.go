package data

import (
	"cyi-app/app/service/marketing/internal/conf"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"cyi-app/app/service/marketing/internal/data/curd"
	mod "cyi-app/app/service/marketing/internal/data/models"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(
	NewCouponRepo,
	NewData,
	NewDB,
)

type Data struct {
}

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

	env := os.Getenv("SERVICE_ENV")
	if env != "production" {
		if err := db.AutoMigrate(&mod.Coupon{}, &mod.CouponDistRecord{}); err != nil {
			panic(err)
		}
	}

	if env == "develop" {
		generateCode(db)
	}

	return db
}

func generateCode(db *gorm.DB) {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./curd",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery,
	})
	g.UseDB(db)
	g.ApplyBasic(
		&mod.Coupon{},
		&mod.CouponDistRecord{},
	)
	g.ApplyInterface(func(mod.ICoupon) {},
		mod.Coupon{},
	)
	g.ApplyInterface(func(mod.ICouponDistRecord) {},
		mod.CouponDistRecord{},
	)
	g.Execute()
}