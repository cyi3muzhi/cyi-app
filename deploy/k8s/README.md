# K8s 部署配置文档

## 目录结构

```
e:\me\cyi-app\deploy\k8s\
├── config/             # 配置文件目录
│   ├── configmap.yaml       # 配置映射文件
│   ├── consul-pvc.yaml      # Consul持久卷声明
│   └── secret.yaml          # 密钥文件
├── monitoring/        # 监控配置目录
│   ├── grafana.yaml         # Grafana监控面板配置
│   └── prometheus.yaml      # Prometheus监控系统配置
└── services/          # 服务配置目录
    ├── deployments.yaml     # 服务部署配置
    └── services.yaml        # 服务网络配置
```

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

#### consul-pvc.yaml
- **功能**：为Consul服务创建持久卷声明
- **内容**：定义了两个PVC，分别用于存储Consul数据和配置
- **作用**：确保Consul的数据在Pod重启后不会丢失，提供持久化存储
- **存储配置**：
  - 每个PVC请求1Gi存储空间
  - 使用ReadWriteOnce访问模式

#### secret.yaml
- **功能**：存储敏感信息
- **内容**：包含数据库密码、Redis密码等敏感配置
- **作用**：以加密方式存储敏感信息，提高安全性
- **主要密钥**：
  - MySQL根密码
  - Redis密码

### 2. monitoring 目录

#### grafana.yaml
- **功能**：部署Grafana监控面板
- **内容**：定义Grafana的Deployment和Service
- **作用**：提供可视化监控界面，展示Prometheus收集的监控数据
- **访问方式**：http://localhost:30300

#### prometheus.yaml
- **功能**：部署Prometheus监控系统
- **内容**：定义Prometheus的Deployment、Service和配置
- **作用**：收集和存储监控数据，为Grafana提供数据源
- **访问方式**：http://localhost:30900

### 3. services 目录

#### deployments.yaml
- **功能**：定义所有服务的部署配置
- **内容**：包含以下服务的部署配置：
  - 基础设施服务：Consul、Redis、MySQL、Jaeger
  - 微服务：service-interface、service-user、service-job、service-task
- **作用**：控制服务的部署方式、副本数量、容器配置等
- **主要配置**：
  - 容器镜像和版本
  - 环境变量
  - 端口映射
  - 资源限制
  - 健康检查

#### services.yaml
- **功能**：定义所有服务的网络配置
- **内容**：为每个服务创建Service，定义服务的访问方式和端口映射
- **作用**：控制服务的网络访问，包括ClusterIP、NodePort等访问方式
- **主要服务**：
  - Consul：NodePort，端口8500:30500
  - Redis：ClusterIP，端口6379
  - MySQL：ClusterIP，端口3306
  - Jaeger：NodePort，端口16686:30686
  - service-interface：NodePort，端口8000:30800、9000:30901
  - service-user：ClusterIP，端口9000
  - service-job：ClusterIP，端口8000
  - service-task：ClusterIP，端口8000

## 部署流程

### 1. 配置准备
```bash
# 应用配置映射
kubectl apply -f config/configmap.yaml

# 应用密钥
kubectl apply -f config/secret.yaml

# 应用Consul持久卷声明
kubectl apply -f config/consul-pvc.yaml
```

### 2. 服务部署
```bash
# 部署服务
kubectl apply -f services/deployments.yaml

# 创建服务网络配置
kubectl apply -f services/services.yaml
```

### 3. 监控部署
```bash
# 部署Prometheus
kubectl apply -f monitoring/prometheus.yaml

# 部署Grafana
kubectl apply -f monitoring/grafana.yaml
```

## 访问方式

| 服务 | 访问地址 | 用途 |
|------|----------|------|
| Consul UI | http://localhost:30500 | 服务注册与配置管理 |
| Grafana | http://localhost:30300 | 监控面板 |
| Prometheus | http://localhost:30900 | 监控数据收集 |
| Jaeger | http://localhost:30686 | 分布式追踪 |
| API接口 | http://localhost:30800 | 通过service-interface访问所有API |

## 服务依赖关系

```
service-interface → service-user → MySQL
                 ↓
          service-job → Redis
                 ↓
          service-task

所有服务 → Consul (服务发现与配置)
所有服务 → Jaeger (分布式追踪)
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

3. **Consul配置问题**：
   - 访问Consul UI：http://localhost:30500
   - 检查KV存储：`kubectl exec -it <consul-pod> -- consul kv get -recurse`

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
