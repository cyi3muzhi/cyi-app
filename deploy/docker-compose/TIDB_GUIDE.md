# TiDB 本地开发环境快速指南

## 🚀 启动 TiDB

### 方式一：单独启动 TiDB
```bash
cd deploy/docker-compose
docker-compose -f docker-compose.tidb.yml up -d
```

### 方式二：查看服务状态
```bash
docker-compose -f docker-compose.tidb.yml ps
```

### 查看日志
```bash
# 查看所有服务日志
docker-compose -f docker-compose.tidb.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.tidb.yml logs -f tidb
```

## 📊 访问 TiDB

### 1. 使用 MySQL 客户端连接
TiDB 默认端口是 4000，用户名为 root，密码为空：
```bash
mysql -h 127.0.0.1 -P 4000 -u root
```

### 2. 创建数据库
```sql
CREATE DATABASE cin_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE cin_test;
```

### 3. 访问 TiDB Dashboard
- PD Dashboard: http://localhost:2379/dashboard
- TiDB Status: http://localhost:10080

## 🔧 连接信息

- **主机**: 127.0.0.1
- **端口**: 4000
- **用户名**: root
- **密码**: (空)
- **默认数据库**: test (可自行创建)

## 📝 DSN 连接字符串示例

Go 应用中使用的 DSN 格式：
```
root:@tcp(127.0.0.1:4000)/cin_test?charset=utf8mb4&parseTime=True&loc=Local
```

## 🛑 停止 TiDB

```bash
cd deploy/docker-compose
docker-compose -f docker-compose.tidb.yml down
```

如果要删除数据卷：
```bash
docker-compose -f docker-compose.tidb.yml down -v
```

## 💡 提示

- TiDB 首次启动需要一些时间来初始化，请耐心等待
- 可以通过 `docker logs tidb-server` 查看启动进度
- 默认没有密码，生产环境请务必设置密码
