package poolconfig

import (
	"database/sql"
	"cyi-app/app/service/user/internal/conf"
	"cyi-app/pkg/util/poolconfig"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"github.com/redis/go-redis/v9"
)

// ExampleNewDB 使用poolconfig配置GORM的示例
func ExampleNewDB(c *conf.Data) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		SkipDefaultTransaction: true,
	})
	if err != nil {
		panic(err)
	}

	// 使用poolconfig配置连接池
	mySQLConfig := poolconfig.GetMySQLConfig()
	
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	
	sqlDB.SetMaxOpenConns(mySQLConfig.MaxOpenConns)
	sqlDB.SetMaxIdleConns(mySQLConfig.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(mySQLConfig.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(mySQLConfig.ConnMaxIdleTime)

	return db
}

// ExampleNewRedisClient 使用poolconfig配置Redis的示例
func ExampleNewRedisClient(c *conf.Data) *redis.Client {
	redisConfig := poolconfig.GetRedisConfig()
	
	client := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Network:      c.Redis.Network,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		Password:     c.Redis.Password,
		// 应用连接池配置
		PoolSize:        redisConfig.PoolSize,
		MinIdleConns:    redisConfig.MinIdleConns,
		MaxIdleConns:    redisConfig.MaxIdleConns,
		ConnMaxLifetime: redisConfig.ConnMaxLifetime,
		ConnMaxIdleTime: redisConfig.ConnMaxIdleTime,
	})

	return client
}
