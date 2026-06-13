# TiDB 最佳实践与优化指南

## 📊 TiDB 分布式特性

### 1. 水平扩展
- **计算扩展**：增加 TiDB Server 节点
- **存储扩展**：增加 TiKV 节点
- 扩展过程对业务透明，无需停机

### 2. 高可用
- 数据默认三副本
- 节点故障自动转移
- Raft 一致性协议保证数据安全

### 3. 分布式事务
- 基于 Google Percolator 模型
- 支持 ACID
- 默认隔离级别：可重复读（REPEATABLE READ）

## 💡 数据库设计最佳实践

### 1. 主键设计
✅ **推荐做法**：
- 使用自增主键（`AUTO_INCREMENT`）
- 使用随机值（如 UUID）作为主键
- 使用 `AUTO_RANDOM`（TiDB 特有功能）

❌ **避免**：
- 使用时间戳作为主键（容易产生热点）
- 使用连续大值作为主键

### 2. 索引设计
- 合理创建索引，避免过多索引
- 使用复合索引优化常用查询
- 定期通过 `EXPLAIN ANALYZE` 分析慢查询

### 3. 分库分表 vs TiDB
- TiDB 自动分片，无需手动分库分表
- 超大表（>1TB）可以考虑分区表

## 🚀 应用开发建议

### 1. 连接池配置
```go
// 示例：配置 SQL 连接池
db.SetMaxOpenConns(200)    // 最大连接数
db.SetMaxIdleConns(10)     // 最大空闲连接数
db.SetConnMaxLifetime(time.Hour)
```

### 2. 事务使用
- 保持事务简短
- 避免大事务（建议 < 100MB 或 < 10 万行）
- 合理选择事务隔离级别

### 3. SQL 优化
- 使用 `EXPLAIN` 分析执行计划
- 避免全表扫描
- 合理使用分页查询

## 🔧 GORM 与 TiDB 配合

你的项目使用 GORM，这是 TiDB 官方推荐的 ORM！

### 不需要修改的部分
- 标准的 CRUD 操作
- 关联查询
- 事务操作

### 可能需要调整的部分
- 某些 MySQL 特有的优化器提示
- 部分特定的分区表语法

## 📈 性能监控

### 关键指标
- QPS（每秒查询数）
- 延迟（TP99、TP95）
- 慢查询
- 存储使用量

### 监控工具
- TiDB Dashboard（http://localhost:2379/dashboard）
- Prometheus + Grafana
- 日志分析

## ⚠️ 常见问题

### 1. TiDB 比 MySQL 慢？
- TiDB 是分布式数据库，网络开销更大
- 适合大数据量场景，小数据量可能 MySQL 更快
- 可以通过优化索引和 SQL 提升性能

### 2. 事务冲突
- 分布式系统中更常见
- 使用乐观事务或悲观事务
- 应用层做好重试逻辑

### 3. 数据迁移
- 使用 Dumpling 导出数据
- 使用 TiDB Lightning 导入数据
- 使用 DM（Data Migration）做增量同步

## 📚 进阶学习资源

- 官方文档：https://docs.pingcap.com/
- TiDB 大学：https://university.pingcap.com/
- GitHub：https://github.com/pingcap/tidb
- AskTUG 社区：https://asktug.com/

## 🎯 下一步建议

1. 在本地启动 TiDB 并连接测试
2. 将测试环境迁移到 TiDB
3. 观察性能，根据监控优化
4. 考虑生产环境部署方案
