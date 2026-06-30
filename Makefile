# Makefile - 项目构建配置文件
# 提供项目级别的构建、部署和代码生成目标

# 系统和环境变量
GOHOSTOS:=$(shell go env GOHOSTOS) # 判断当前系统是否为 Windows
GOPATH:=$(shell go env GOPATH)    # 获取 GOPATH 环境变量
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "dev") # 获取当前 Git 仓库的版本标签



# 应用相对路径
APP_RELATIVE_PATH=$(notdir $(patsubst %/,%,$(dir $(CURDIR))))/$(notdir $(CURDIR))

# 跨平台 proto 文件查找
# 直接硬编码 API_PROTO_FILES 和 INTERNAL_PROTO_FILES 的值
# 这样可以确保路径格式正确，避免 PowerShell 命令执行错误
INTERNAL_PROTO_FILES=
API_PROTO_FILES=api/blades/v1/blades.proto api/cin/admin/user/user.proto api/cin/interface/v1/cin_interface.proto api/cin/interface/v1/cin_interface_error.proto api/job/v1/job.proto api/service/order/v1/order.proto api/service/product/v1/product.proto api/service/user/v1/user.proto api/service/user/v1/user_error.proto api/task/v1/task.proto

# 服务发现
SERVICE_CMD_DIRS := $(wildcard app/*/*/cmd/* app/*/cmd/*) # 查找所有服务的 cmd 目录
SERVICE_NAMES := $(foreach dir,$(SERVICE_CMD_DIRS),$(notdir $(dir))) # 从 cmd 目录中提取服务名称

# 为每个服务生成 wire 代码
ifeq ($(OS),Windows_NT)
    # Windows PowerShell 语法
    $(foreach service,$(SERVICE_NAMES),$(eval .PHONY: wire-$(service)))
    $(foreach service,$(SERVICE_NAMES),$(eval wire-$(service):; cd $(subst /,\,$(filter %/cmd/$(service),$(SERVICE_CMD_DIRS))); wire))
else
    # bash 语法
    $(foreach service,$(SERVICE_NAMES),$(eval .PHONY: wire-$(service)))
    $(foreach service,$(SERVICE_NAMES),$(eval wire-$(service):; cd $(filter %/cmd/$(service),$(SERVICE_CMD_DIRS)) && wire))
endif

.PHONY: debug-wire
# 调试 wire 命令
debug-wire:
	@echo "SERVICE_NAMES=$(SERVICE_NAMES)"
	@echo "SERVICE_CMD_DIRS=$(SERVICE_CMD_DIRS)"
	@echo "OS=$(OS)"
	@echo "Current directory: $(CURDIR)"

# 目标定义

.PHONY: init
# 初始化环境变量和依赖
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@1.28.1
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@1.28.1
	go install github.com/go-kratos/kratos/cmd/kratos/v2@2.8.3
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@1.28.1
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
# 生成 internal 目录下的 proto 文件
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# 生成 api 目录下的 proto 文件
api:
	protoc --proto_path=./ \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./api \
	       --go-http_out=paths=source_relative:./api \
	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
		   --go-vtproto_out=paths=source_relative:./api \
    	   --go-vtproto_opt=features=marshal+unmarshal+size \
	       $(API_PROTO_FILES)

.PHONY: validate
# 生成 validate 目录下的 proto 文件
validate:
	protoc --proto_path=. \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:. \
	       --validate_out=paths=source_relative,lang=go:. \
	       $(API_PROTO_FILES)

.PHONY: errors
# 生成 errors 目录下的 proto 文件
errors:
	protoc --proto_path=. \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:. \
	       --go-errors_out=paths=source_relative:. \
	       $(API_PROTO_FILES)

.PHONY: build
# 构建项目
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: generate
# 生成项目代码
generate:
	go generate ./...
	go mod tidy

.PHONY: all
# 生成所有 proto 文件并构建项目
all:
	@$(MAKE) api
	@$(MAKE) config
	@$(MAKE) generate

.PHONY: service
# 生成服务模板
service:
	@powershell -Command "$$goals = '$(MAKECMDGOALS)'.Split(' '); $$args = $$goals | Where-Object { $$_ -ne 'service' }; if ($$args.Count -eq 0) { echo 'Error: Service name is required. Usage: make service <service-name> [options]'; echo 'Options: --description=<description> --version=<version> --author=<author> --features=<features>'; exit 1; } cd tools/service-gen; go run main.go $$args"



.PHONY: wire
# 生成 wire 代码
wire:
	$(foreach service,$(SERVICE_NAMES),wire-$(service))

.PHONY: docker-up
# 运行所有 docker 容器（先启动基础设施，再启动应用）
docker-up: docker-up-infra docker-up-app

.PHONY: docker-up-infra
# 运行基础设施 docker 容器
docker-up-infra:
	cd deploy/docker-compose && docker-compose -f docker-compose.infra.yml -p cyi-app up -d

.PHONY: docker-up-app
# 运行应用 docker 容器
docker-up-app:
	cd deploy/docker-compose && docker-compose -f docker-compose.app.yml -p cyi-app up -d

.PHONY: docker-down
# 停止所有 docker 容器
docker-down: docker-down-app docker-down-infra

.PHONY: docker-down-infra
# 停止基础设施 docker 容器
docker-down-infra:
	cd deploy/docker-compose && docker-compose -f docker-compose.infra.yml -p cyi-app down

.PHONY: docker-down-app
# 停止应用 docker 容器
docker-down-app:
	cd deploy/docker-compose && docker-compose -f docker-compose.app.yml -p cyi-app down

# 为每个服务生成 docker-build 目标
$(foreach service,$(SERVICE_NAMES),$(eval docker-build-$(service):; cd $(patsubst %/cmd/$(service),%,$(filter %/cmd/$(service),$(SERVICE_CMD_DIRS))) && make docker))

.PHONY: docker-build
# 构建 docker 镜像
docker-build:
	$(foreach service,$(SERVICE_NAMES),$(MAKE) docker-build-$(service))

.PHONY: k8s-up
# 部署 Kubernetes 服务
k8s-up:
	cd deploy/k8s && powershell -ExecutionPolicy Bypass -File .\deploy-all.ps1

.PHONY: k8s-down
# 停止 Kubernetes 服务
k8s-down:
	kubectl delete deployment --all
	kubectl delete service --all
	kubectl delete daemonset --all

.PHONY: k8s-clean
# 清理 Kubernetes 资源
k8s-clean:
	kubectl delete deployment --all
	kubectl delete service --all
	kubectl delete pvc --all
	kubectl delete configmap --all
	kubectl delete secret --all
	kubectl delete daemonset --all

.PHONY: env
# 显示环境变量
env:
	@echo 'GOHOSTOS: $(GOHOSTOS)'
	@echo 'GOPATH: $(GOPATH)'
	@echo 'VERSION: $(VERSION)'
	@echo 'APP_RELATIVE_PATH: $(APP_RELATIVE_PATH)'
	@echo 'SERVICE_CMD_DIRS: $(SERVICE_CMD_DIRS)'
	@echo 'SERVICE_NAMES: $(SERVICE_NAMES)'
	@echo 'API_PROTO_FILES: $(API_PROTO_FILES)'

.PHONY: help
# 显示帮助信息
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.PHONY: debug
# 调试信息
debug:
	@echo 'GOHOSTOS=$(GOHOSTOS)'
	@echo 'API_PROTO_FILES=$(API_PROTO_FILES)'
	@echo 'INTERNAL_PROTO_FILES=$(INTERNAL_PROTO_FILES)'
	@echo 'SERVICE_NAMES=$(SERVICE_NAMES)'
	@echo 'SERVICE_CMD_DIRS=$(SERVICE_CMD_DIRS)'

.DEFAULT_GOAL := help

# 忽略传递给 service 目标的参数
%:
	@: