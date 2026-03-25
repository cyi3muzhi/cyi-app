# Kubernetes 配置文件详细说明

本文档详细解释了 `e:\me\cyi-app\deploy\k8s` 目录下所有配置文件的内容和用途。

## 1. 根目录文件

### README.md
**用途**：Kubernetes 部署配置文档，包含目录结构、部署流程、访问方式等说明

### deploy-all.ps1
**用途**：一键部署脚本，用于快速部署整个系统

## 2. config 目录

### configmap.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: ConfigMap  # 资源类型：配置映射
metadata:
  name: cyi-app-config  # ConfigMap 名称
  namespace: default  # 命名空间
data:
  # 数据库配置
  MYSQL_HOST: "mysql"  # MySQL 主机名
  MYSQL_PORT: "3306"  # MySQL 端口
  MYSQL_DATABASE: "cyi_app"  # 数据库名称
  
  # Redis 配置
  REDIS_HOST: "redis"  # Redis 主机名
  REDIS_PORT: "6379"  # Redis 端口
  
  # Jaeger 配置
  JAEGER_ADDR: "jaeger:6831"  # Jaeger 服务地址
  
  # Consul 配置
  CONSUL_ADDR: "consul:8500"  # Consul 服务地址
  
  # 服务环境
  SERVICE_ENV: "dev"  # 服务环境：开发环境
```
**用途**：存储应用配置的键值对，将配置与应用代码分离，方便配置管理和更新

### secret.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: Secret  # 资源类型：密钥
metadata:
  name: cyi-app-secret  # Secret 名称
  namespace: default  # 命名空间
type: Opaque  # 密钥类型：不透明（自定义）
data:
  # 数据库密码（base64 编码）
  MYSQL_ROOT_PASSWORD: "cXdlc3QxMjM="  # MySQL 根密码：qwest123
  
  # Redis 密码（base64 编码）
  REDIS_PASSWORD: "cXdlc3QxMjM="  # Redis 密码：qwest123
```
**用途**：存储敏感信息，以加密方式存储敏感配置，提高安全性

## 3. infrastructure 目录

### consul 目录

#### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: consul  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: consul  # 选择标签
  template:
    metadata:
      labels:
        app: consul  # Pod 标签
    spec:
      containers:
      - name: consul  # 容器名称
        image: consul:1.22.0  # 容器镜像
        ports:
        - containerPort: 8500  # HTTP 端口
        - containerPort: 8301  # LAN 端口
        - containerPort: 8302  # WAN 端口
        - containerPort: 8600  # DNS 端口
        command:  # 启动命令
        - consul
        - agent
        - -server
        - -bootstrap-expect=1
        - -ui
        - -bind=0.0.0.0
        - -client=0.0.0.0
        volumeMounts:
        - name: consul-data  # 数据卷
          mountPath: /consul/data
        - name: consul-config  # 配置卷
          mountPath: /consul/config
      volumes:
      - name: consul-data
        persistentVolumeClaim:
          claimName: consul-data-pvc  # 数据持久卷声明
      - name: consul-config
        persistentVolumeClaim:
          claimName: consul-config-pvc  # 配置持久卷声明
```
**用途**：部署 Consul 服务发现和配置管理服务

#### pvc.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: PersistentVolumeClaim  # 资源类型：持久卷声明
metadata:
  name: consul-data-pvc  # 数据卷声明名称
  namespace: default  # 命名空间
spec:
  accessModes:
  - ReadWriteOnce  # 访问模式：单节点读写
  resources:
    requests:
      storage: 1Gi  # 存储请求：1GB
---
apiVersion: v1  # Kubernetes API 版本
kind: PersistentVolumeClaim  # 资源类型：持久卷声明
metadata:
  name: consul-config-pvc  # 配置卷声明名称
  namespace: default  # 命名空间
spec:
  accessModes:
  - ReadWriteOnce  # 访问模式：单节点读写
  resources:
    requests:
      storage: 1Gi  # 存储请求：1GB
```
**用途**：为 Consul 服务创建持久卷声明，确保数据和配置持久化

#### service.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: consul  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: consul  # 选择标签
  ports:
  - name: http  # 端口名称
    port: 8500  # 服务端口
    targetPort: 8500  # 容器端口
    nodePort: 30500  # 节点端口（外部访问）
  - name: dns  # 端口名称
    port: 8600  # 服务端口
    targetPort: 8600  # 容器端口
  type: NodePort  # 服务类型：NodePort
```
**用途**：为 Consul 服务创建网络访问配置，提供外部访问端口

### elk 目录

#### elasticsearch 目录

##### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: elasticsearch  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: elasticsearch  # 选择标签
  template:
    metadata:
      labels:
        app: elasticsearch  # Pod 标签
    spec:
      containers:
      - name: elasticsearch  # 容器名称
        image: docker.elastic.co/elasticsearch/elasticsearch:7.17.10  # 容器镜像
        ports:
        - containerPort: 9200  # HTTP 端口
        - containerPort: 9300  # 节点间通信端口
        env:
        - name: discovery.type  # 发现类型
          value: "single-node"  # 单节点模式
        - name: ES_JAVA_OPTS  # JAVA 选项
          value: "-Xms512m -Xmx512m"  # 内存配置
        volumeMounts:
        - name: elasticsearch-data  # 数据卷
          mountPath: /usr/share/elasticsearch/data
      volumes:
      - name: elasticsearch-data
        persistentVolumeClaim:
          claimName: elasticsearch-pvc  # 数据持久卷声明
```
**用途**：部署 Elasticsearch 服务，用于存储和索引日志数据

##### pvc.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: PersistentVolumeClaim  # 资源类型：持久卷声明
metadata:
  name: elasticsearch-pvc  # 卷声明名称
  namespace: default  # 命名空间
spec:
  accessModes:
  - ReadWriteOnce  # 访问模式：单节点读写
  resources:
    requests:
      storage: 10Gi  # 存储请求：10GB
```
**用途**：为 Elasticsearch 服务创建持久卷声明，确保数据持久化

#### filebeat 目录

##### config.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: ConfigMap  # 资源类型：配置映射
metadata:
  name: filebeat-config  # ConfigMap 名称
  namespace: default  # 命名空间
data:
  filebeat.yml: |  # Filebeat 配置
    # 容器日志文件输入
    filebeat.inputs:
      # 所有服务日志
      - type: log
        enabled: true
        paths:
          - /app/logs/*/*.log  # 日志文件路径
        encoding: utf-8  # 编码格式
        multiline.pattern: '^\d{4}-\d{2}-\d{2}'  # 多行日志模式
        multiline.negate: true  # 否定模式
        multiline.match: after  # 匹配方式
        processors:
          - dissect:  # 解析处理器
              tokenizer: '%{log_dir}/%{filename}'  # 分词器
              field: 'log.file.path'  # 字段
              target_prefix: 'parsed'  # 目标前缀
          - rename:  # 重命名处理器
              fields:
                parsed.filename: 'log_filename'  # 重命名字段
          - add_fields:  # 添加字段处理器
              fields:
                service: app  # 服务名称
                environment: ${ENVIRONMENT:development}  # 环境变量
              target: ''  # 目标
        fields_under_root: true  # 字段在根级别

    # 输出配置
    output.logstash:
      hosts: ["logstash:5044"]  # Logstash 地址
      ssl.enabled: false  # 禁用 SSL

    # 日志配置
    logging.level: info  # 日志级别
    logging.to_files: false  # 不输出到文件
```
**用途**：配置 Filebeat 日志收集器，定义日志收集和处理规则

##### daemonset.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: DaemonSet  # 资源类型：守护进程集
metadata:
  name: filebeat  # 守护进程集名称
  namespace: default  # 命名空间
  labels:
    app: filebeat  # 标签
spec:
  selector:
    matchLabels:
      app: filebeat  # 选择标签
  template:
    metadata:
      labels:
        app: filebeat  # Pod 标签
    spec:
      containers:
      - name: filebeat  # 容器名称
        image: docker.elastic.co/beats/filebeat:7.17.10  # 容器镜像
        args:
        - -e  # 启用日志
        - --strict.perms=false  # 禁用严格权限
        env:
        - name: ENVIRONMENT  # 环境变量
          value: development  # 环境值
        volumeMounts:
        - name: filebeat-config  # 配置卷
          mountPath: /usr/share/filebeat/filebeat.yml
          subPath: filebeat.yml
        - name: filebeat-data  # 数据卷
          mountPath: /usr/share/filebeat/data
        - name: logs  # 日志卷
          mountPath: /app/logs  # 挂载路径
      volumes:
      - name: filebeat-config
        configMap:
          name: filebeat-config  # 配置映射名称
      - name: filebeat-data
        emptyDir: {}  # 临时目录
      - name: logs
        hostPath:
          path: /mnt/e/me/cyi-app/deploy/logs  # 主机路径
          type: DirectoryOrCreate  # 目录不存在则创建
```
**用途**：部署 Filebeat 守护进程，在每个节点上运行日志收集器

#### kibana 目录

##### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: kibana  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: kibana  # 选择标签
  template:
    metadata:
      labels:
        app: kibana  # Pod 标签
    spec:
      containers:
      - name: kibana  # 容器名称
        image: docker.elastic.co/kibana/kibana:7.17.10  # 容器镜像
        ports:
        - containerPort: 5601  # 端口
        env:
        - name: ELASTICSEARCH_HOSTS  # Elasticsearch 地址
          value: "http://elasticsearch:9200"  # 地址值
---
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: kibana  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: kibana  # 选择标签
  ports:
  - port: 5601  # 服务端口
    targetPort: 5601  # 容器端口
    nodePort: 30201  # 节点端口（外部访问）
  type: NodePort  # 服务类型：NodePort
```
**用途**：部署 Kibana 服务，用于可视化和分析日志数据

#### logstash 目录

##### config.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: ConfigMap  # 资源类型：配置映射
metadata:
  name: logstash-config  # ConfigMap 名称
  namespace: default  # 命名空间
data:
  logstash.conf: |  # Logstash 配置
    input {
      beats {
        port => 5044  # Beats 输入端口
      }
    }

    filter {
      # 解析 JSON 格式日志
      json {
        source => "message"  # 消息源
        target => "log"  # 目标字段
        remove_field => ["message"]  # 移除原始消息字段
      }

      # 处理时间戳
      date {
        match => ["log.time", "yyyy-MM-dd HH:mm:ss"]  # 时间格式
        target => "@timestamp"  # 目标时间戳
        timezone => "Asia/Shanghai"  # 时区
      }

      # 提取服务名称
      if [log_filename] {
        grok {
          match => { "log_filename" => "dev\.(?<service>[^-]+)" }  # 匹配服务名称
        }
      }
    }

    output {
      elasticsearch {
        hosts => ["elasticsearch:9200"]  # Elasticsearch 地址
        index => "cyi-%{[service]}-%{+YYYY.MM.dd}"  # 索引名称
      }
    }
```
**用途**：配置 Logstash 日志处理器，定义日志处理和输出规则

##### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: logstash  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: logstash  # 选择标签
  template:
    metadata:
      labels:
        app: logstash  # Pod 标签
    spec:
      containers:
      - name: logstash  # 容器名称
        image: docker.elastic.co/logstash/logstash:7.17.10  # 容器镜像
        ports:
        - containerPort: 5044  # Beats 输入端口
        volumeMounts:
        - name: logstash-config  # 配置卷
          mountPath: /usr/share/logstash/pipeline
      volumes:
      - name: logstash-config
        configMap:
          name: logstash-config  # 配置映射名称
---
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: logstash  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: logstash  # 选择标签
  ports:
  - port: 5044  # 服务端口
    targetPort: 5044  # 容器端口
```
**用途**：部署 Logstash 服务，用于处理和转换日志数据

### jaeger 目录

#### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: jaeger  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: jaeger  # 选择标签
  template:
    metadata:
      labels:
        app: jaeger  # Pod 标签
    spec:
      containers:
      - name: jaeger  # 容器名称
        image: jaegertracing/all-in-one:1.47  # 容器镜像
        ports:
        - containerPort: 16686  # UI 端口
        - containerPort: 6831  # UDP 端口
        - containerPort: 4318  # OTLP HTTP 端口
        volumeMounts:
        - name: jaeger-data  # 数据卷
          mountPath: /jaeger/data
      volumes:
      - name: jaeger-data
        persistentVolumeClaim:
          claimName: jaeger-pvc  # 数据持久卷声明
---
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: jaeger  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: jaeger  # 选择标签
  ports:
  - port: 16686  # 服务端口
    targetPort: 16686  # 容器端口
    nodePort: 30686  # 节点端口（外部访问）
  - port: 6831  # 服务端口
    targetPort: 6831  # 容器端口
    protocol: UDP  # 协议
  - port: 4318  # 服务端口
    targetPort: 4318  # 容器端口
  type: NodePort  # 服务类型：NodePort
```
**用途**：部署 Jaeger 分布式追踪服务，用于追踪分布式系统中的请求流程

#### pvc.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: PersistentVolumeClaim  # 资源类型：持久卷声明
metadata:
  name: jaeger-pvc  # 卷声明名称
  namespace: default  # 命名空间
spec:
  accessModes:
  - ReadWriteOnce  # 访问模式：单节点读写
  resources:
    requests:
      storage: 5Gi  # 存储请求：5GB
```
**用途**：为 Jaeger 服务创建持久卷声明，确保追踪数据持久化

#### service.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: jaeger  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: jaeger  # 选择标签
  ports:
  - name: ui  # 端口名称
    port: 16686  # 服务端口
    targetPort: 16686  # 容器端口
    nodePort: 30686  # 节点端口（外部访问）
  - name: zipkin  # 端口名称
    port: 9411  # 服务端口
    targetPort: 9411  # 容器端口
  - name: otlp-http  # 端口名称
    port: 4318  # 服务端口
    targetPort: 4318  # 容器端口
  type: NodePort  # 服务类型：NodePort
```
**用途**：为 Jaeger 服务创建网络访问配置，提供外部访问端口

### mysql 目录

#### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: mysql  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: mysql  # 选择标签
  template:
    metadata:
      labels:
        app: mysql  # Pod 标签
    spec:
      containers:
      - name: mysql  # 容器名称
        image: mysql:8.0  # 容器镜像
        ports:
        - containerPort: 3306  # 端口
        env:
        - name: MYSQL_ROOT_PASSWORD  # 根密码
          valueFrom:
            secretKeyRef:
              name: cyi-app-secret  # 密钥名称
              key: MYSQL_ROOT_PASSWORD  # 密钥键
        - name: MYSQL_DATABASE  # 数据库名称
          value: "cyi_app"
        volumeMounts:
        - name: mysql-data  # 数据卷
          mountPath: /var/lib/mysql
      volumes:
      - name: mysql-data
        persistentVolumeClaim:
          claimName: mysql-pvc  # 数据持久卷声明
```
**用途**：部署 MySQL 数据库服务，为应用提供持久化数据存储

#### pvc.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: PersistentVolumeClaim  # 资源类型：持久卷声明
metadata:
  name: mysql-pvc  # 卷声明名称
  namespace: default  # 命名空间
spec:
  accessModes:
  - ReadWriteOnce  # 访问模式：单节点读写
  resources:
    requests:
      storage: 10Gi  # 存储请求：10GB
```
**用途**：为 MySQL 服务创建持久卷声明，确保数据库数据持久化

#### service.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: mysql  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: mysql  # 选择标签
  ports:
  - port: 3306  # 服务端口
    targetPort: 3306  # 容器端口
  type: ClusterIP  # 服务类型：ClusterIP
```
**用途**：为 MySQL 服务创建网络访问配置，提供集群内部访问

### redis 目录

#### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: redis  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: redis  # 选择标签
  template:
    metadata:
      labels:
        app: redis  # Pod 标签
    spec:
      containers:
      - name: redis  # 容器名称
        image: redis:7.0  # 容器镜像
        ports:
        - containerPort: 6379  # 端口
        env:
        - name: REDIS_PASSWORD  # 密码
          valueFrom:
            secretKeyRef:
              name: cyi-app-secret  # 密钥名称
              key: REDIS_PASSWORD  # 密钥键
        command:  # 启动命令
        - redis-server
        - --requirepass
        - $(REDIS_PASSWORD)
        volumeMounts:
        - name: redis-data  # 数据卷
          mountPath: /data
      volumes:
      - name: redis-data
        persistentVolumeClaim:
          claimName: redis-pvc  # 数据持久卷声明
```
**用途**：部署 Redis 缓存服务，为应用提供缓存功能

#### pvc.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: PersistentVolumeClaim  # 资源类型：持久卷声明
metadata:
  name: redis-pvc  # 卷声明名称
  namespace: default  # 命名空间
spec:
  accessModes:
  - ReadWriteOnce  # 访问模式：单节点读写
  resources:
    requests:
      storage: 5Gi  # 存储请求：5GB
```
**用途**：为 Redis 服务创建持久卷声明，确保缓存数据持久化

#### service.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: redis  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: redis  # 选择标签
  ports:
  - port: 6379  # 服务端口
    targetPort: 6379  # 容器端口
  type: ClusterIP  # 服务类型：ClusterIP
```
**用途**：为 Redis 服务创建网络访问配置，提供集群内部访问

## 4. microservices 目录

### interface 目录

#### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: service-interface  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: service-interface  # 选择标签
  template:
    metadata:
      labels:
        app: service-interface  # Pod 标签
    spec:
      containers:
      - name: service-interface  # 容器名称
        image: go-kratos/cin-interface:0.1.0  # 容器镜像
        ports:
        - containerPort: 8000  # HTTP 端口
        - containerPort: 9000  # gRPC 端口
        env:
        - name: CONSUL_ADDR  # Consul 地址
          value: "consul:8500"
        - nameCONSUL_PATHS  # Consul 路径
          value: "/develop/cyi-app/cin/interface"
        - name: SERVICE_ENV  # 服务环境
          value: "dev"
        - name: LOG_LEVEL  # 日志级别
          value: "debug"
        volumeMounts:
        - name: logs  # 日志卷
          mountPath: /app/logs  # 挂载路径
      volumes:
      - name: logs
        hostPath:
          path: /mnt/e/me/cyi-app/deploy/logs/interface  # 主机路径
          type: DirectoryOrCreate  # 目录不存在则创建
```
**用途**：部署接口服务，作为 API 网关，处理外部请求并路由到其他服务

#### service.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: service-interface  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: service-interface  # 选择标签
  ports:
  - name: http  # 端口名称
    port: 8000  # 服务端口
    targetPort: 8000  # 容器端口
    nodePort: 30800  # 节点端口（外部访问）
  - name: grpc  # 端口名称
    port: 9000  # 服务端口
    targetPort: 9000  # 容器端口
    nodePort: 30901  # 节点端口（外部访问）
  type: NodePort  # 服务类型：NodePort
```
**用途**：为接口服务创建网络访问配置，提供外部访问端口

### job 目录

#### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: service-job  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: service-job  # 选择标签
  template:
    metadata:
      labels:
        app: service-job  # Pod 标签
    spec:
      containers:
      - name: service-job  # 容器名称
        image: go-kratos/cin-job:0.1.0  # 容器镜像
        ports:
        - containerPort: 8000  # 端口
        env:
        - name: CONSUL_ADDR  # Consul 地址
          value: "consul:8500"
        - nameCONSUL_PATHS  # Consul 路径
          value: "/develop/cyi-app/cin/job"
        - name: SERVICE_ENV  # 服务环境
          value: "dev"
        - name: LOG_LEVEL  # 日志级别
          value: "debug"
        volumeMounts:
        - name: logs  # 日志卷
          mountPath: /app/logs  # 挂载路径
      volumes:
      - name: logs
        hostPath:
          path: /mnt/e/me/cyi-app/deploy/logs/job  # 主机路径
          type: DirectoryOrCreate  # 目录不存在则创建
```
**用途**：部署任务服务，处理定时任务和后台任务

#### service.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: service-job  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: service-job  # 选择标签
  ports:
  - port: 8000  # 服务端口
    targetPort: 8000  # 容器端口
  type: ClusterIP  # 服务类型：ClusterIP
```
**用途**：为任务服务创建网络访问配置，提供集群内部访问

### task 目录

#### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: service-task  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: service-task  # 选择标签
  template:
    metadata:
      labels:
        app: service-task  # Pod 标签
    spec:
      containers:
      - name: service-task  # 容器名称
        image: go-kratos/cin-task:0.1.0  # 容器镜像
        ports:
        - containerPort: 8000  # 端口
        env:
        - name: CONSUL_ADDR  # Consul 地址
          value: "consul:8500"
        - nameCONSUL_PATHS  # Consul 路径
          value: "/develop/cyi-app/cin/task"
        - name: SERVICE_ENV  # 服务环境
          value: "dev"
        - name: LOG_LEVEL  # 日志级别
          value: "debug"
        volumeMounts:
        - name: logs  # 日志卷
          mountPath: /app/logs  # 挂载路径
      volumes:
      - name: logs
        hostPath:
          path: /mnt/e/me/cyi-app/deploy/logs/task  # 主机路径
          type: DirectoryOrCreate  # 目录不存在则创建
```
**用途**：部署任务处理服务，处理具体的任务执行逻辑

#### service.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: service-task  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: service-task  # 选择标签
  ports:
  - port: 8000  # 服务端口
    targetPort: 8000  # 容器端口
  type: ClusterIP  # 服务类型：ClusterIP
```
**用途**：为任务处理服务创建网络访问配置，提供集群内部访问

### user 目录

#### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: service-user  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: service-user  # 选择标签
  template:
    metadata:
      labels:
        app: service-user  # Pod 标签
    spec:
      containers:
      - name: service-user  # 容器名称
        image: go-kratos/cin-user:0.1.0  # 容器镜像
        ports:
        - containerPort: 9000  # 端口
        env:
        - name: CONSUL_ADDR  # Consul 地址
          value: "consul:8500"
        - nameCONSUL_PATHS  # Consul 路径
          value: "/develop/cyi-app/cin/user"
        - name: SERVICE_ENV  # 服务环境
          value: "dev"
        - name: LOG_LEVEL  # 日志级别
          value: "debug"
        volumeMounts:
        - name: logs  # 日志卷
          mountPath: /app/logs  # 挂载路径
      volumes:
      - name: logs
        hostPath:
          path: /mnt/e/me/cyi-app/deploy/logs/user  # 主机路径
          type: DirectoryOrCreate  # 目录不存在则创建
```
**用途**：部署用户服务，处理用户相关的业务逻辑

#### service.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: service-user  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: service-user  # 选择标签
  ports:
  - port: 9000  # 服务端口
    targetPort: 9000  # 容器端口
  type: ClusterIP  # 服务类型：ClusterIP
```
**用途**：为用户服务创建网络访问配置，提供集群内部访问

## 5. monitoring 目录

### grafana 目录

#### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: grafana  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: grafana  # 选择标签
  template:
    metadata:
      labels:
        app: grafana  # Pod 标签
    spec:
      containers:
      - name: grafana  # 容器名称
        image: grafana/grafana:9.5.2  # 容器镜像
        ports:
        - containerPort: 3000  # 端口
        volumeMounts:
        - name: grafana-data  # 数据卷
          mountPath: /var/lib/grafana
      volumes:
      - name: grafana-data
        persistentVolumeClaim:
          claimName: grafana-pvc  # 数据持久卷声明
---
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: grafana  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: grafana  # 选择标签
  ports:
  - port: 3000  # 服务端口
    targetPort: 3000  # 容器端口
    nodePort: 30300  # 节点端口（外部访问）
  type: NodePort  # 服务类型：NodePort
```
**用途**：部署 Grafana 监控面板，可视化监控数据

#### pvc.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: PersistentVolumeClaim  # 资源类型：持久卷声明
metadata:
  name: grafana-pvc  # 卷声明名称
  namespace: default  # 命名空间
spec:
  accessModes:
  - ReadWriteOnce  # 访问模式：单节点读写
  resources:
    requests:
      storage: 5Gi  # 存储请求：5GB
```
**用途**：为 Grafana 服务创建持久卷声明，确保配置和数据持久化

### prometheus 目录

#### deployment.yaml
```yaml
apiVersion: apps/v1  # Kubernetes API 版本
kind: Deployment  # 资源类型：部署
metadata:
  name: prometheus  # 部署名称
  namespace: default  # 命名空间
spec:
  replicas: 1  # 副本数
  selector:
    matchLabels:
      app: prometheus  # 选择标签
  template:
    metadata:
      labels:
        app: prometheus  # Pod 标签
    spec:
      containers:
      - name: prometheus  # 容器名称
        image: prom/prometheus:v2.43.0  # 容器镜像
        ports:
        - containerPort: 9090  # 端口
        volumeMounts:
        - name: prometheus-config  # 配置卷
          mountPath: /etc/prometheus
        - name: prometheus-data  # 数据卷
          mountPath: /prometheus
      volumes:
      - name: prometheus-config
        configMap:
          name: prometheus-config  # 配置映射名称
      - name: prometheus-data
        persistentVolumeClaim:
          claimName: prometheus-pvc  # 数据持久卷声明
---
apiVersion: v1  # Kubernetes API 版本
kind: Service  # 资源类型：服务
metadata:
  name: prometheus  # 服务名称
  namespace: default  # 命名空间
spec:
  selector:
    app: prometheus  # 选择标签
  ports:
  - port: 9090  # 服务端口
    targetPort: 9090  # 容器端口
    nodePort: 30900  # 节点端口（外部访问）
  type: NodePort  # 服务类型：NodePort
---
apiVersion: v1  # Kubernetes API 版本
kind: ConfigMap  # 资源类型：配置映射
metadata:
  name: prometheus-config  # ConfigMap 名称
  namespace: default  # 命名空间
data:
  prometheus.yml: |  # Prometheus 配置
    global:
      scrape_interval: 15s  # 抓取间隔

    scrape_configs:
      - job_name: 'kubernetes-pods'  # 作业名称
        kubernetes_sd_configs:
        - role: pod  # 角色：Pod

        relabel_configs:
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
          action: keep
          regex: true  # 保持带有 prometheus.io/scrape: true 注解的 Pod
        - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
          action: replace
          target_label: __metrics_path__
          regex: (.+)
        - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
          action: replace
          regex: ([^:]+)(?::\d+)?;(\d+)
          replacement: $1:$2
          target_label: __address__
        - action: labelmap
          regex: __meta_kubernetes_pod_label_(.+)
        - source_labels: [__meta_kubernetes_namespace]
          action: replace
          target_label: kubernetes_namespace
        - source_labels: [__meta_kubernetes_pod_name]
          action: replace
          target_label: kubernetes_pod_name
```
**用途**：部署 Prometheus 监控系统，收集和存储监控数据

#### pvc.yaml
```yaml
apiVersion: v1  # Kubernetes API 版本
kind: PersistentVolumeClaim  # 资源类型：持久卷声明
metadata:
  name: prometheus-pvc  # 卷声明名称
  namespace: default  # 命名空间
spec:
  accessModes:
  - ReadWriteOnce  # 访问模式：单节点读写
  resources:
    requests:
      storage: 10Gi  # 存储请求：10GB
```
**用途**：为 Prometheus 服务创建持久卷声明，确保监控数据持久化

## 配置文件用途总结

| 配置文件 | 用途 |
|---------|------|
| config/configmap.yaml | 存储应用配置的键值对，如数据库连接信息、服务地址等 |
| config/secret.yaml | 存储敏感信息，如数据库密码、Redis密码等 |
| infrastructure/consul/deployment.yaml | 部署 Consul 服务发现和配置管理服务 |
| infrastructure/consul/pvc.yaml | 为 Consul 服务创建持久卷声明 |
| infrastructure/consul/service.yaml | 为 Consul 服务创建网络访问配置 |
| infrastructure/elk/elasticsearch/deployment.yaml | 部署 Elasticsearch 服务，用于存储和索引日志数据 |
| infrastructure/elk/elasticsearch/pvc.yaml | 为 Elasticsearch 服务创建持久卷声明 |
| infrastructure/elk/filebeat/config.yaml | 配置 Filebeat 日志收集器 |
| infrastructure/elk/filebeat/daemonset.yaml | 部署 Filebeat 守护进程，在每个节点上运行日志收集器 |
| infrastructure/elk/kibana/deployment.yaml | 部署 Kibana 服务，用于可视化和分析日志数据 |
| infrastructure/elk/logstash/config.yaml | 配置 Logstash 日志处理器 |
| infrastructure/elk/logstash/deployment.yaml | 部署 Logstash 服务，用于处理和转换日志数据 |
| infrastructure/jaeger/deployment.yaml | 部署 Jaeger 分布式追踪服务 |
| infrastructure/jaeger/pvc.yaml | 为 Jaeger 服务创建持久卷声明 |
| infrastructure/jaeger/service.yaml | 为 Jaeger 服务创建网络访问配置 |
| infrastructure/mysql/deployment.yaml | 部署 MySQL 数据库服务 |
| infrastructure/mysql/pvc.yaml | 为 MySQL 服务创建持久卷声明 |
| infrastructure/mysql/service.yaml | 为 MySQL 服务创建网络访问配置 |
| infrastructure/redis/deployment.yaml | 部署 Redis 缓存服务 |
| infrastructure/redis/pvc.yaml | 为 Redis 服务创建持久卷声明 |
| infrastructure/redis/service.yaml | 为 Redis 服务创建网络访问配置 |
| microservices/interface/deployment.yaml | 部署接口服务，作为 API 网关 |
| microservices/interface/service.yaml | 为接口服务创建网络访问配置 |
| microservices/job/deployment.yaml | 部署任务服务，处理定时任务和后台任务 |
| microservices/job/service.yaml | 为任务服务创建网络访问配置 |
| microservices/task/deployment.yaml | 部署任务处理服务，处理具体的任务执行逻辑 |
| microservices/task/service.yaml | 为任务处理服务创建网络访问配置 |
| microservices/user/deployment.yaml | 部署用户服务，处理用户相关的业务逻辑 |
| microservices/user/service.yaml | 为用户服务创建网络访问配置 |
| monitoring/grafana/deployment.yaml | 部署 Grafana 监控面板 |
| monitoring/grafana/pvc.yaml | 为 Grafana 服务创建持久卷声明 |
| monitoring/prometheus/deployment.yaml | 部署 Prometheus 监控系统 |
| monitoring/prometheus/pvc.yaml | 为 Prometheus 服务创建持久卷声明 |
| deploy-all.ps1 | 一键部署脚本，用于快速部署整个系统 |
| README.md | Kubernetes 部署配置文档，包含目录结构、部署流程、访问方式等说明 |
