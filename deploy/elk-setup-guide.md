
## 前置条件

1. **Docker** - 用于运行 ELK Stack 容器
2. **Docker Compose** - 用于管理多容器应用
3. **Filebeat** - 用于收集和发送日志（使用 Docker 运行）

## 部署步骤

### 步骤 1: 启动 ELK Stack

1. **进入部署目录**：
   ```powershell
   cd e:\me\cyi-app\deploy
   ```

2. **启动 ELK Stack**：
   ```powershell
   docker-compose -f elk-docker-compose.yml up -d
   ```

3. **验证服务状态**：
   ```powershell
   docker-compose -f elk-docker-compose.yml ps
   ```

   确保所有服务都处于 `Up` 状态。

### 步骤 2: 配置并启动 Filebeat

1. **创建 Filebeat Docker Compose 文件**：
   ```yaml
   # filebeat-docker-compose.yml
   version: '3.8'
   
   services:
     filebeat:
       image: docker.elastic.co/beats/filebeat:7.17.10
       container_name: cyi-app-filebeat
       volumes:
         - ./filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
         - e:/me/cyi-app/app/service/user/cmd/user/logs:/logs/user-service:ro
         - e:/me/cyi-app/app/cin/interface/cmd/interface/logs:/logs/interface-service:ro
       networks:
         - deploy_default
       command: filebeat -e -strict.perms=false
   
   networks:
     deploy_default:
       external: true
   ```

2. **创建 Filebeat 配置文件**：
   ```yaml
   # filebeat.yml
   filebeat.inputs:
   # 用户服务日志
   - type: log
     enabled: true
     paths:
       - /logs/user-service/*.log
     encoding: utf-8
     multiline.pattern: '^\d{4}-\d{2}-\d{2}'
     multiline.negate: true
     multiline.match: after
     fields:
       service: user-service
       environment: development
   
   # 接口服务日志
   - type: log
     enabled: true
     paths:
       - /logs/interface-service/*.log
     encoding: utf-8
     multiline.pattern: '^\d{4}-\d{2}-\d{2}'
     multiline.negate: true
     multiline.match: after
     fields:
       service: interface-service
       environment: development
   
   output.elasticsearch:
     hosts: ["cyi-app-elasticsearch:9200"]
     index: "cyi-app-logs-%{+yyyy.MM.dd}"
   
   setup.template.name: "cyi-app-logs"
   setup.template.pattern: "cyi-app-logs-*"
   setup.template.enabled: true
   setup.template.overwrite: true
   
   logging.level: info
   logging.to_files: true
   logging.files:
     path: /usr/share/filebeat/logs
     name: filebeat.log
     keepfiles: 7
     permissions: 0644
   ```

3. **启动 Filebeat**：
   ```powershell
   docker-compose -f filebeat-docker-compose.yml up -d
   ```

### 步骤 3: 验证日志流

1. **生成测试日志**：
   运行您的应用，生成一些日志。

2. **检查 Elasticsearch 索引**：
   ```powershell
   Invoke-WebRequest -Uri "http://localhost:9200/_cat/indices" -UseBasicParsing
   ```

   您应该看到类似 `cyi-app-logs-2026.01.31` 的索引。

3. **检查 Filebeat 日志**：
   ```powershell
   docker logs cyi-app-filebeat
   ```

   确保没有错误信息。

4. **检查 Logstash 日志**：
   ```powershell
   docker logs cyi-app-logstash
   ```

   确保没有错误信息。

### 步骤 4: 在 Kibana 中查看日志

1. **访问 Kibana**：
   打开浏览器，访问 http://localhost:5601

2. **登录 Kibana**：
   - 用户名：`elastic`
   - 密码：`changeme`

3. **创建索引模式**：
   - 点击左侧菜单中的 "Stack Management" → "Index Patterns"
   - 点击 "Create index pattern"
   - 输入索引模式：`cyi-app-logs-*`
   - 点击 "Next step"
   - 选择时间字段：`@timestamp`
   - 点击 "Create index pattern"

4. **查看日志**：
   - 点击左侧菜单中的 "Analytics" → "Discover"
   - 选择您刚刚创建的索引模式
   - 您应该能看到应用生成的日志
   - 使用 `service` 字段过滤不同微服务的日志

## 多微服务日志配置说明

### 配置原理

本配置使用了方法一：为每个微服务配置单独的输入，具体包括：

1. **独立的卷映射**：
   - 用户服务：`e:/me/cyi-app/app/service/user/cmd/user/logs:/logs/user-service:ro`
   - 接口服务：`e:/me/cyi-app/app/cin/interface/cmd/interface/logs:/logs/interface-service:ro`

2. **独立的输入配置**：
   - 为每个微服务创建单独的 `filebeat.inputs` 配置
   - 为每个微服务添加唯一的 `service` 字段
   - 保持统一的日志处理规则

### 优点

- **灵活性高**：每个微服务可以有独立的配置和字段
- **易于识别**：通过 `service` 字段可以清晰区分不同微服务的日志
- **易于过滤**：在 Kibana 中可以通过 `service` 字段快速过滤特定微服务的日志
- **易于扩展**：添加新微服务时只需添加新的卷映射和输入配置

### 如何添加新微服务

1. **修改 filebeat-docker-compose.yml**：
   添加新的卷映射：
   ```yaml
   volumes:
     - ./filebeat.yml:/usr/share/filebeat/filebeat.yml:ro
     - e:/me/cyi-app/app/service/user/cmd/user/logs:/logs/user-service:ro
     - e:/me/cyi-app/app/cin/interface/cmd/interface/logs:/logs/interface-service:ro
     - e:/me/cyi-app/path/to/new-service/logs:/logs/new-service:ro  # 新微服务
   ```

2. **修改 filebeat.yml**：
   添加新的输入配置：
   ```yaml
   # 新微服务日志
   - type: log
     enabled: true
     paths:
       - /logs/new-service/*.log
     encoding: utf-8
     multiline.pattern: '^\d{4}-\d{2}-\d{2}'
     multiline.negate: true
     multiline.match: after
     fields:
       service: new-service
       environment: development
   ```

3. **重启 Filebeat**：
   ```powershell
   docker-compose -f filebeat-docker-compose.yml down
   docker-compose -f filebeat-docker-compose.yml up -d
   ```

## 配置说明

### Filebeat 配置 (`filebeat.yml`)

- **用户服务日志**：监控 `/logs/user-service/*.log` 路径
- **接口服务日志**：监控 `/logs/interface-service/*.log` 路径
- **多行处理**：处理堆栈信息等多行日志
- **自定义字段**：为每个微服务添加唯一的 `service` 字段
- **输出配置**：发送到 Elasticsearch 容器

### Filebeat Docker Compose 配置 (`filebeat-docker-compose.yml`)

- **卷映射**：
  - `./filebeat.yml:/usr/share/filebeat/filebeat.yml:ro` - 配置文件映射
  - `e:/me/cyi-app/app/service/user/cmd/user/logs:/logs/user-service:ro` - 用户服务日志目录映射
  - `e:/me/cyi-app/app/cin/interface/cmd/interface/logs:/logs/interface-service:ro` - 接口服务日志目录映射
- **网络配置**：使用与 ELK Stack 相同的网络，确保容器间通信
- **命令配置**：启用详细输出，禁用严格权限检查

### Logstash 配置 (`logstash.conf`)

- **输入**：接收来自 Filebeat 的日志
- **过滤**：
  - 解析 JSON 格式的日志
  - 解析文本格式的日志（使用 grok）
  - 处理日期字段
  - 添加服务字段
  - 移除不需要的字段
- **输出**：发送到 Elasticsearch 并输出到控制台（用于调试）

### ELK Stack 配置 (`elk-docker-compose.yml`)

- **Elasticsearch**：单节点部署，内存限制为 1GB
- **Logstash**：接收来自 Filebeat 的日志，内存限制为 512MB
- **Kibana**：连接到 Elasticsearch，提供 Web 界面

## 常见问题及解决方案

### 1. Filebeat 找不到日志文件

**症状**：Filebeat 日志中显示 "No such file or directory" 错误

**解决方案**：
- 检查主机上的实际日志文件路径
- 确保 Docker 卷映射正确
- 确保日志文件存在且可读

### 2. Elasticsearch 没有索引

**症状**：`Invoke-WebRequest -Uri "http://localhost:9200/_cat/indices" -UseBasicParsing` 没有显示 `cyi-app-logs-*` 索引

**解决方案**：
- 检查 Filebeat 日志中的错误信息：`docker logs cyi-app-filebeat`
- 确保 Filebeat 配置中的 Elasticsearch 主机正确：`cyi-app-elasticsearch:9200`
- 确保 Filebeat 与 Elasticsearch 在同一网络中
- 运行应用程序生成一些测试日志

### 3. Kibana 看不到索引

**症状**：Kibana 显示 "Ready to try Kibana? First, you need data."

**解决方案**：
- 确保 Elasticsearch 中存在 `cyi-app-logs-*` 索引
- 在 Kibana 中创建正确的索引模式
- 检查时间范围设置，确保包含日志生成的时间

### 4. Filebeat 启动失败

**症状**：`docker-compose -f filebeat-docker-compose.yml up -d` 失败

**解决方案**：
- 检查 Docker 网络是否存在：`docker network ls`
- 确保 ELK Stack 已经启动：`docker-compose -f elk-docker-compose.yml ps`
- 检查 Filebeat 配置文件格式是否正确

### 5. 日志格式解析错误

**症状**：日志显示为原始文本，未被正确解析

**解决方案**：
- 检查 Logstash 管道配置中的 grok 模式
- 确保日志格式与配置中的模式匹配
- 查看 Logstash 日志中的解析错误

## 故障排除步骤

### 步骤 1: 检查所有服务状态

```powershell
# 检查 ELK Stack 服务状态
docker-compose -f elk-docker-compose.yml ps

# 检查 Filebeat 服务状态
docker-compose -f filebeat-docker-compose.yml ps
```

### 步骤 2: 检查服务日志

```powershell
# 查看 Elasticsearch 日志
docker logs cyi-app-elasticsearch

# 查看 Logstash 日志
docker logs cyi-app-logstash

# 查看 Kibana 日志
docker logs cyi-app-kibana

# 查看 Filebeat 日志
docker logs cyi-app-filebeat
```

### 步骤 3: 验证 Elasticsearch 索引

```powershell
# 检查所有索引
Invoke-WebRequest -Uri "http://localhost:9200/_cat/indices" -UseBasicParsing

# 检查特定索引
Invoke-WebRequest -Uri "http://localhost:9200/cyi-app-logs-*/_count" -UseBasicParsing
```

### 步骤 4: 验证日志文件存在

```powershell
# 查看用户服务日志文件
Get-ChildItem -Path "e:\me\cyi-app\app\service\user\cmd\user\logs" -Name

# 查看接口服务日志文件
Get-ChildItem -Path "e:\me\cyi-app\app\cin\interface\cmd\interface\logs" -Name
```

### 步骤 5: 重新配置并重启服务

```powershell
# 停止所有服务
docker-compose -f filebeat-docker-compose.yml down
docker-compose -f elk-docker-compose.yml down

# 重新启动 ELK Stack
docker-compose -f elk-docker-compose.yml up -d

# 等待服务启动
Start-Sleep -Seconds 30

# 重新启动 Filebeat
docker-compose -f filebeat-docker-compose.yml up -d
```

## 高级配置

### 1. 启用身份验证

**修改 Elasticsearch 密码**：
在 `elk-docker-compose.yml` 中修改 `ELASTIC_PASSWORD` 环境变量。

**更新 Filebeat 配置**：
```yaml
output.elasticsearch:
  hosts: ["cyi-app-elasticsearch:9200"]
  username: "elastic"
  password: "your-new-password"
```

### 2. 配置 HTTPS

**生成证书**：
使用 Elasticsearch 内置的证书工具生成 SSL 证书。

**更新配置**：
- 修改 `elk-docker-compose.yml` 中的相关配置
- 更新 `filebeat.yml` 中的输出配置，启用 HTTPS

### 3. 调整内存限制

根据系统资源情况，调整 `elk-docker-compose.yml` 中的 JVM 内存设置：
- `ES_JAVA_OPTS`：Elasticsearch 内存限制
- `LS_JAVA_OPTS`：Logstash 内存限制

## 性能优化

1. **增加 Elasticsearch 内存**：对于生产环境，建议至少分配 4GB 内存
2. **使用多个 Filebeat 实例**：对于高流量应用，考虑使用多个 Filebeat 实例
3. **调整 Logstash 管道**：根据日志量调整工作线程和批处理大小
4. **启用索引生命周期管理**：自动管理旧索引，防止存储空间不足

## 监控 ELK Stack

1. **使用 Elasticsearch 监控**：访问 http://localhost:9200/_cluster/health
2. **使用 Kibana 监控**：在 Kibana 中查看 "Monitoring" 页面
3. **使用 Filebeat 监控**：监控 ELK Stack 自身的日志

## 清理

### 停止 ELK Stack

```powershell
docker-compose -f elk-docker-compose.yml down
```

### 停止 Filebeat

```powershell
docker-compose -f filebeat-docker-compose.yml down
```

### 清理数据卷

```powershell
docker volume rm deploy_es_data deploy_logstash_config deploy_kibana_config
```

## 总结

通过本指南，您已经成功将多个微服务的日志集成到 ELK Stack 中，实现了：

- **多微服务日志集中管理**：所有微服务日志统一存储在 Elasticsearch 中
- **微服务日志清晰区分**：通过 `service` 字段可以清晰区分不同微服务的日志
- **日志实时分析**：通过 Logstash 实时处理和转换日志
- **日志可视化**：在 Kibana 中查看和分析日志
- **日志告警**：可配置基于日志内容的告警

ELK Stack 为您的多微服务架构提供了强大的日志管理和分析能力，帮助您快速定位和解决问题，提高系统可靠性和可维护性。