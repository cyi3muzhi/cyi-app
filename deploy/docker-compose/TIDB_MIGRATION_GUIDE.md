# TiDB 测试与迁移完整指南

---

## 🚀 第一阶段：本地测试环境搭建

### 1.1 启动方式

```bash
# 进入目录
cd e:\me\cyi-app\deploy\docker-compose

# 启动TiDB（后台模式）
docker-compose -f docker-compose.tidb.yml up -d

# 查看状态
docker-compose -f docker-compose.tidb.yml ps

# 查看日志
docker-compose -f docker-compose.tidb.yml logs -f
```

### 1.2 验证服务

```bash
# 检查PD状态
curl http://localhost:2379/pd/api/v1/health

# 检查TiDB状态
curl http://localhost:10080/status

# 使用MySQL客户端连接
mysql -h 127.0.0.1 -P 4000 -u root
```

### 1.3 创建测试数据库

```sql
-- 创建数据库
CREATE DATABASE cin_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户（可选）
CREATE USER 'app_user'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON cin_test.* TO 'app_user'@'%';
FLUSH PRIVILEGES;
```

---

## 📦 第二阶段：数据迁移

### 2.1 方案一：使用 mysqldump + mysqlimport

```bash
# 从MySQL导出数据
mysqldump -h mysql_host -P 3306 -u root -p cin_test > backup.sql

# 导入到TiDB
mysql -h 127.0.0.1 -P 4000 -u root cin_test < backup.sql
```

### 2.2 方案二：使用TiDB官方工具（推荐）

#### Dumpling（数据导出）
```bash
# 安装Dumpling
tiup install dumpling

# 导出数据
tiup dumpling \
  -h mysql_host \
  -P 3306 \
  -u root \
  -p password \
  -d cin_test \
  -o ./backup
```

#### TiDB Lightning（数据导入）
```bash
# 安装TiDB Lightning
tiup install lightning

# 创建配置文件 tidb-lightning.toml
cat > tidb-lightning.toml <<EOF
[lightning]
# 日志
level = "info"
file = "tidb-lightning.log"

[mydumper]
# 数据源目录
data-source-dir = "./backup"

[tidb]
# TiDB连接信息
host = "127.0.0.1"
port = 4000
user = "root"
password = ""
status-port = 10080
EOF

# 启动导入
tiup lightning -config tidb-lightning.toml
```

### 2.3 方案三：使用DM（Data Migration）增量同步

适用于需要在线迁移的场景：

```bash
# 创建DM配置文件
cat > dm.toml <<EOF
[global]
worker-count = 4
log-level = "info"

[source.mysql]
host = "mysql_host"
port = 3306
user = "root"
password = "password"

[target.tidb]
host = "127.0.0.1"
port = 4000
user = "root"
password = ""

[task.migrate]
source-id = "mysql"
target-database = "cin_test"
table-list = ["cin_test.*"]
EOF

# 启动DM
tiup dmctl --master-addr pd:2379 start-task dm.toml
```

---

## 📊 第三阶段：性能测试与监控

### 3.1 连接应用测试

修改配置文件 `config.yaml`：

```yaml
data:
  database:
    driver: mysql
    source: root:@tcp(127.0.0.1:4000)/cin_test?parseTime=True&loc=Local&charset=utf8mb4
```

### 3.2 运行测试

```bash
# 启动应用服务
cd e:\me\cyi-app\app\service\order
go run main.go

# 运行单元测试
go test -v ./...
```

### 3.3 监控工具

#### TiDB Dashboard
- **地址**: http://localhost:2379/dashboard
- **功能**:
  - 集群状态监控
  - 慢查询分析
  - 事务冲突检测
  - 性能指标查看

#### Prometheus + Grafana
```yaml
# 在docker-compose中添加监控
prometheus:
  image: prom/prometheus:latest
  volumes:
    - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
  ports:
    - "9090:9090"

grafana:
  image: grafana/grafana:latest
  ports:
    - "3000:3000"
  environment:
    - GF_INSTALL_PLUGINS=grafana-piechart-panel
```

### 3.4 性能优化建议

| 优化项 | 说明 |
|--------|------|
| **索引优化** | 使用`EXPLAIN ANALYZE`分析慢查询，添加合适索引 |
| **批量操作** | 使用批量插入/更新减少网络往返 |
| **连接池配置** | 根据并发量调整连接池大小 |
| **分区表** | 对超大表使用分区表优化查询 |
| **TiFlash** | 为分析查询创建TiFlash副本 |

---

## 🎯 第四阶段：生产环境部署方案

### 4.1 硬件建议

| 角色 | CPU | 内存 | 存储 |
|------|-----|------|------|
| TiDB Server | 16+ cores | 32+ GB | SSD 200+ GB |
| PD Server | 8+ cores | 16+ GB | SSD 100+ GB |
| TiKV Server | 16+ cores | 32+ GB | NVMe SSD 1TB+ |

### 4.2 拓扑规划

```
生产环境推荐配置（最小规模）：
├── PD: 3节点（奇数，保证高可用）
├── TiKV: 3节点（每节点3副本，共9副本）
└── TiDB: 2节点（负载均衡）

大型生产环境配置：
├── PD: 5节点
├── TiKV: 6-12节点（根据数据量扩展）
├── TiDB: 3-5节点（根据QPS扩展）
└── TiFlash: 2-3节点（可选，用于分析）
```

### 4.3 使用TiUP部署（推荐）

```bash
# 安装TiUP
curl --proto '=https' --tlsv1.2 -sSf https://tiup-mirrors.pingcap.com/install.sh | sh

# 初始化集群配置
tiup cluster template > topology.yaml

# 编辑配置，设置节点信息
vi topology.yaml

# 部署集群
tiup cluster deploy mytidb v7.5.0 ./topology.yaml --user root -p

# 启动集群
tiup cluster start mytidb

# 查看集群状态
tiup cluster display mytidb
```

### 4.4 Kubernetes部署

使用TiDB Operator：

```yaml
# 安装TiDB Operator
helm repo add pingcap https://charts.pingcap.org/
helm install tidb-operator pingcap/tidb-operator --namespace tidb-admin

# 部署TiDB集群
kubectl apply -f tidb-cluster.yaml
```

### 4.5 高可用配置

```yaml
# TiKV副本配置
replication:
  max-replicas: 3        # 每Region的副本数
  location-labels:       # 跨可用区部署
    - zone
    - rack
    - host

# PD调度策略
scheduler:
  high-availability: true
  balance:
    leader: true
    region: true
```

---

## 📝 检查清单

- [ ] TiDB服务启动正常
- [ ] 数据迁移完成且验证通过
- [ ] 应用连接测试通过
- [ ] 性能测试达到预期
- [ ] 监控告警配置完成
- [ ] 备份恢复测试完成
- [ ] 故障转移测试完成

---

## 📚 参考资料

- TiDB官方文档：https://docs.pingcap.com/
- TiUP文档：https://docs.pingcap.com/tiup/stable
- DM文档：https://docs.pingcap.com/tidb-data-migration/stable
- 性能优化：https://docs.pingcap.com/tidb/stable/performance-tuning
