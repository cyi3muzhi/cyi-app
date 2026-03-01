# K8s 部署配置文档

## 目录结构

```
e:\me\cyi-app\deploy\k8s\
├── config/             # 全局配置文件目录
│   ├── configmap.yaml  # 配置映射文件
│   └── secret.yaml     # 密钥文件
├── infrastructure/     # 基础设施服务目录
│   ├── consul/         # Consul服务发现配置
│   │   ├── deployment.yaml
│   │   ├── pvc.yaml
│   │   └── service.yaml
│   ├── elk/            # ELK日志收集堆栈
│   │   ├── elasticsearch/
│   │   │   ├── deployment.yaml
│   │   │   └── pvc.yaml
│   │   ├── filebeat/
│   │   │   ├── config.yaml
│   │   │   └── daemonset.yaml
│   │   ├── kibana/
│   │   │   └── deployment.yaml
│   │   └── logstash/
│   │       ├── config.yaml
│   │       └── deployment.yaml
│   ├── jaeger/         # 分布式追踪服务
│   │   ├── deployment.yaml
│   │   ├── pvc.yaml
│   │   └── service.yaml
│   ├── mysql/          # 数据库服务
│   │   ├── deployment.yaml
│   │   ├── pvc.yaml
│   │   └── service.yaml
│   └── redis/          # 缓存服务
│       ├── deployment.yaml
│       ├── pvc.yaml
│       └── service.yaml
├── microservices/      # 微服务目录
│   ├── interface/      # 接口服务
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── job/            # 任务服务
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   ├── task/           # 任务处理服务
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   └── user/           # 用户服务
│       ├── deployment.yaml
│       └── service.yaml
├── monitoring/         # 监控系统目录
│   ├── grafana/        # 监控可视化
│   │   ├── deployment.yaml
│   │   └── pvc.yaml
│   └── prometheus/     # 监控数据收集
│       ├── deployment.yaml
│       └── pvc.yaml
├── README.md           # 本文档
└── deploy-all.ps1      # 一键部署脚本
```

## 设计理念

### 目录结构设计原因

1. **模块化组织**：
   - 将不同类型的服务和配置分离到独立目录，便于管理和维护
   - 每个组件拥有自己的子目录，包含所有相关配置文件
   - 遵循"关注点分离"原则，使配置结构清晰易懂

2. **层次化架构**：
   - 从底层基础设施到上层微服务，形成清晰的层次结构
   - 基础设施服务作为底层支撑，微服务作为业务逻辑实现
   - 监控系统独立部署，便于统一监控所有服务

3. **可扩展性**：
   - 新服务可以轻松添加到相应目录，无需修改整体结构
   - 每个服务的配置文件集中管理，便于快速定位和修改
   - 支持服务的独立部署和扩展

4. **标准化命名**：
   - 统一使用 `deployment.yaml`、`service.yaml`、`pvc.yaml` 等标准化文件名
   - 配置文件命名直观，便于理解其功能
   - 遵循 Kubernetes 最佳实践，提高配置的可读性

## 各文件功能详细介绍

### 1. config 目录

#### configmap.yaml
- **功能**：存储应用配置的键值对
- **内容**：包含数据库连接信息、服务地址、环境变量等配置
- **作用**：将配置与应用代码分离，方便配置管理和更新
- **主要配置项**：
  - 数据库连接信息（MySQL主机、端口、数据库名）
  - Redis连接信息（主机、端口）
  - Jaeger跟踪服务地址
  - Consul服务地址
  - 服务环境变量

#### secret.yaml
- **功能**：存储敏感信息
- **内容**：包含数据库密码、Redis密码等敏感配置
- **作用**：以加密方式存储敏感信息，提高安全性
- **主要密钥**：
  - MySQL根密码
  - Redis密码

### 2. infrastructure 目录

#### consul 目录
- **功能**：部署Consul服务发现和配置管理
- **内容**：包含Consul的Deployment、Service和PVC配置
- **作用**：提供服务注册、发现和配置管理功能
- **访问方式**：http://localhost:30500

#### elk 目录
- **功能**：部署ELK日志收集堆栈
- **组件**：
  - **elasticsearch**：存储和索引日志数据
  - **logstash**：处理和转换日志
  - **kibana**：可视化日志数据
  - **filebeat**：收集容器日志
- **作用**：集中管理和分析所有服务的日志
- **访问方式**：
  - Kibana：http://localhost:30201
  - Elasticsearch：http://localhost:30200

#### jaeger 目录
- **功能**：部署分布式追踪服务
- **内容**：包含Jaeger的Deployment、Service和PVC配置
- **作用**：追踪分布式系统中的请求流程，便于故障排查
- **访问方式**：http://localhost:30686

#### mysql 目录
- **功能**：部署MySQL数据库服务
- **内容**：包含MySQL的Deployment、Service和PVC配置
- **作用**：为应用提供持久化数据存储
- **访问方式**：ClusterIP，端口3306

#### redis 目录
- **功能**：部署Redis缓存服务
- **内容**：包含Redis的Deployment、Service和PVC配置
- **作用**：为应用提供缓存服务，提高性能
- **访问方式**：ClusterIP，端口6379

### 3. microservices 目录

#### interface 目录
- **功能**：部署接口服务
- **内容**：包含接口服务的Deployment和Service配置
- **作用**：作为API网关，处理外部请求并路由到其他服务
- **访问方式**：http://localhost:30800

#### user 目录
- **功能**：部署用户服务
- **内容**：包含用户服务的Deployment和Service配置
- **作用**：处理用户相关的业务逻辑
- **访问方式**：ClusterIP，端口9000

#### job 目录
- **功能**：部署任务服务
- **内容**：包含任务服务的Deployment和Service配置
- **作用**：处理定时任务和后台任务
- **访问方式**：ClusterIP，端口8000

#### task 目录
- **功能**：部署任务处理服务
- **内容**：包含任务处理服务的Deployment和Service配置
- **作用**：处理具体的任务执行逻辑
- **访问方式**：ClusterIP，端口8000

### 4. monitoring 目录

#### prometheus 目录
- **功能**：部署Prometheus监控系统
- **内容**：包含Prometheus的Deployment和PVC配置
- **作用**：收集和存储监控数据
- **访问方式**：http://localhost:30900

#### grafana 目录
- **功能**：部署Grafana监控面板
- **内容**：包含Grafana的Deployment和PVC配置
- **作用**：可视化监控数据，提供监控面板
- **访问方式**：http://localhost:30300

## 部署流程

### 1. 一键部署

使用提供的部署脚本可以快速部署整个系统：

```powershell
# 进入k8s目录
cd e:\me\cyi-app\deploy\k8s

# 执行部署脚本
./deploy-all.ps1
```

### 2. 分步部署

如果需要分步部署，可以按照以下顺序执行：

#### 1. 应用配置
```powershell
# 应用配置映射
kubectl apply -f config/configmap.yaml

# 应用密钥
kubectl apply -f config/secret.yaml
```

#### 2. 部署基础设施服务
```powershell
# 部署Consul
kubectl apply -f infrastructure/consul/

# 部署Redis
kubectl apply -f infrastructure/redis/

# 部署MySQL
kubectl apply -f infrastructure/mysql/

# 部署Jaeger
kubectl apply -f infrastructure/jaeger/

# 部署ELK stack
kubectl apply -f infrastructure/elk/elasticsearch/
kubectl apply -f infrastructure/elk/logstash/
kubectl apply -f infrastructure/elk/kibana/
kubectl apply -f infrastructure/elk/filebeat/
```

#### 3. 部署微服务
```powershell
# 部署接口服务
kubectl apply -f microservices/interface/

# 部署用户服务
kubectl apply -f microservices/user/

# 部署任务服务
kubectl apply -f microservices/job/

# 部署任务处理服务
kubectl apply -f microservices/task/
```

#### 4. 部署监控系统
```powershell
# 部署Prometheus
kubectl apply -f monitoring/prometheus/

# 部署Grafana
kubectl apply -f monitoring/grafana/
```

## 访问方式

| 服务 | 访问地址 | 用途 |
|------|----------|------|
| Consul UI | http://localhost:30500 | 服务注册与配置管理 |
| Kibana | http://localhost:30201 | 日志可视化 |
| Elasticsearch | http://localhost:30200 | 日志存储和索引 |
| Jaeger | http://localhost:30686 | 分布式追踪 |
| Grafana | http://localhost:30300 | 监控面板 |
| Prometheus | http://localhost:30900 | 监控数据收集 |
| API接口 | http://localhost:30800 | 通过interface服务访问所有API |

## 服务依赖关系

```
service-interface → service-user → MySQL
                 ↓
          service-job → Redis
                 ↓
          service-task

所有服务 → Consul (服务发现与配置)
所有服务 → Jaeger (分布式追踪)
所有服务 → ELK (日志收集)
所有服务 → Prometheus (监控)
```

## 故障排查

1. **服务无法启动**：
   - 检查Pod状态：`kubectl get pods`
   - 查看Pod日志：`kubectl logs <pod-name>`
   - 检查配置是否正确：`kubectl get configmap` 和 `kubectl get secret`

2. **服务无法访问**：
   - 检查Service状态：`kubectl get services`
   - 检查网络策略：`kubectl get networkpolicy`
   - 检查Pod网络：`kubectl exec -it <pod-name> -- ping <service-name>`

3. **日志收集问题**：
   - 检查Filebeat状态：`kubectl get pods | Select-String filebeat`
   - 查看Filebeat日志：`kubectl logs <filebeat-pod>`
   - 检查Elasticsearch索引：`Invoke-WebRequest -Uri http://localhost:30200/_cat/indices?v`

4. **监控问题**：
   - 检查Prometheus状态：`kubectl get pods | Select-String prometheus`
   - 检查Grafana状态：`kubectl get pods | Select-String grafana`

## 扩展建议

1. **高可用部署**：
   - 增加服务副本数
   - 使用StatefulSet部署有状态服务
   - 配置PodDisruptionBudget

2. **监控增强**：
   - 配置Grafana告警
   - 添加更多监控指标
   - 集成Alertmanager

3. **安全加固**：
   - 配置网络策略
   - 使用RBAC控制访问
   - 定期更新密钥

4. **CI/CD集成**：
   - 配置Kubernetes资源的版本管理
   - 集成GitOps工作流
   - 自动化部署流程

5. **性能优化**：
   - 配置资源请求和限制
   - 使用HorizontalPodAutoscaler自动缩放
   - 优化存储配置

## 技术栈说明

- **容器编排**：Kubernetes
- **服务发现**：Consul
- **数据存储**：MySQL、Redis
- **日志管理**：ELK Stack (Elasticsearch, Logstash, Kibana, Filebeat)
- **监控系统**：Prometheus + Grafana
- **分布式追踪**：Jaeger
- **部署工具**：PowerShell脚本

本部署架构采用了现代微服务架构的最佳实践，提供了一个完整、可扩展的应用部署方案。