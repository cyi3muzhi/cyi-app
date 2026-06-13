# Pool Config
连接池配置优化工具

## 功能说明

这个工具包提供了基于CPU核心数自动优化的MySQL和Redis连接池配置。

## 安装

```go
import "cyi-app/pkg/util/poolconfig"
```

## 使用示例

### 获取MySQL连接池配置

```go
import "cyi-app/pkg/util/poolconfig"

// 获取默认配置（CPU核心数 * 2
cfg := poolconfig.GetMySQLConfig()

// 自定义乘数（例如 CPU * 3）
cfg := poolconfig.GetMySQLConfig(3)

// 输出配置
fmt.Println(cfg)
```

### 获取Redis连接池配置

```go
import "cyi-app/pkg/util/poolconfig"

// 获取默认配置（CPU核心数 * 5
cfg := poolconfig.GetRedisConfig()

// 自定义乘数
cfg := poolconfig.GetRedisConfig(10)

fmt.Println(cfg)
```

### 在GORM中使用

```go
import (
    "cyi-app/pkg/util/poolconfig"
)

func NewDB(c *conf.Data) *gorm.DB {
    db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
        // ... 配置
    })
    
    // 获取配置
    mySQLConfig := poolconfig.GetMySQLConfig()
    
    // 应用配置
    sqlDB, err := db.DB()
    sqlDB.SetMaxOpenConns(mySQLConfig.MaxOpenConns)
    sqlDB.SetMaxIdleConns(mySQLConfig.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(mySQLConfig.ConnMaxLifetime)
    sqlDB.SetConnMaxIdleTime(mySQLConfig.ConnMaxIdleTime)
    
    return db
}
```

### 在Redis客户端中使用

```go
import (
    "cyi-app/pkg/util/poolconfig"
)

func NewRedisClient(c *conf.Data) *redis.Client {
    // 获取配置
    redisConfig := poolconfig.GetRedisConfig()
    
    client := redis.NewClient(&redis.Options{
        Addr:         c.Redis.Addr,
        // ... 其他配置
        // 应用连接池配置
        PoolSize:        redisConfig.PoolSize,
        MinIdleConns:    redisConfig.MinIdleConns,
        MaxIdleConns:    redisConfig.MaxIdleConns,
        ConnMaxLifetime: redisConfig.ConnMaxLifetime,
        ConnMaxIdleTime: redisConfig.ConnMaxIdleTime,
    })
    
    return client
}
```

## 配置说明

### MySQL配置

| 参数 | 说明 | 默认值 |
|------|------|--------|
| MaxOpenConns | 最大打开连接数 | CPU核心数 * 2 |
| MaxIdleConns | 最大空闲连接数 | MaxOpenConns / 2 (最小5) |
| ConnMaxLifetime | 连接最大存活时间 | 1小时 |
| ConnMaxIdleTime | 连接最大空闲时间 | 10分钟 |

### Redis配置

| 参数 | 说明 | 默认值 |
|------|------|--------|
| PoolSize | 连接池大小 | CPU核心数 * 5 |
| MinIdleConns | 最小空闲连接数 | PoolSize / 4 (最小5) |
| MaxIdleConns | 最大空闲连接数 | PoolSize / 2 (最小10) |
| ConnMaxLifetime | 连接最大存活时间 | 1小时 |
| ConnMaxIdleTime | 连接最大空闲时间 | 10分钟 |

## API

### GetMySQLConfig(multiplier ...int) *MySQLConfig
获取推荐的MySQL连接池配置。

### GetRedisConfig(multiplier ...int) *RedisConfig
获取推荐的Redis连接池配置。

### GetCPUCount() int
获取CPU核心数。

## 注意事项

- MySQL最大空闲连接数默认设置为最大打开连接数的一半，且最小为5
- Redis连接池大小默认是CPU核心数的5倍
- 连接最大存活时间默认1小时，最大空闲时间10分钟
