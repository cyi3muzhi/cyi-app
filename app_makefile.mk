# app_makefile.mk - 应用构建配置文件
# 提供通用的构建、部署和代码生成目标

GOPATH:=$(shell go env GOPATH)

# 获取版本号（跨平台兼容）
VERSION=$(shell git describe --tags --always 2>/dev/null || echo "dev")

# 获取当前目录和父目录名称（跨平台兼容）
CURRENT_DIR=$(notdir $(CURDIR))
PARENT_DIR=$(notdir $(patsubst %/,%,$(dir $(CURDIR))))

# 计算 APP_RELATIVE_PATH：如果父目录是 app，只使用当前目录名称
ifeq ($(PARENT_DIR),app)
    APP_RELATIVE_PATH=$(CURRENT_DIR)
else
    APP_RELATIVE_PATH=$(PARENT_DIR)/$(CURRENT_DIR)
endif

# 根据 app_makefile 位置定位项目根目录
# 获取 app_makefile 文件的绝对路径
APP_MAKEFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
# 获取 app_makefile 所在目录（即项目根目录）
PROJECT_ROOT := $(dir $(APP_MAKEFILE_PATH))
# 移除末尾的斜杠
PROJECT_ROOT := $(patsubst %/,%,$(PROJECT_ROOT))
UNIX_PROJECT_ROOT=$(subst \,/,$(PROJECT_ROOT))

# 计算从当前目录到项目根目录的相对路径
# 规则：基于 APP_RELATIVE_PATH 中的目录数量计算
# 例如：APP_RELATIVE_PATH=blades → 需要 ../../
# 例如：APP_RELATIVE_PATH=service/user → 需要 ../../../
RELATIVE_PATH_TO_ROOT=../../
ifneq ($(findstring /,$(APP_RELATIVE_PATH)),)
    RELATIVE_PATH_TO_ROOT=../../../
endif

# 获取 API proto 文件（跨平台兼容）
ifeq ($(OS),Windows_NT)
    # Windows 平台使用 Git Bash 执行 find 命令
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git)))) 
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find ${RELATIVE_PATH_TO_ROOT}api/${APP_RELATIVE_PATH} -name '*.proto'")  
    INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name '*.proto'")  
else
    # macOS/Linux 平台直接使用 bash 执行 find 命令
    API_PROTO_FILES=$(shell find ${UNIX_PROJECT_ROOT}/api/${APP_RELATIVE_PATH} -name '*.proto')  
    INTERNAL_PROTO_FILES=$(shell find internal -name '*.proto')  
endif

# 获取 Kratos 版本（跨平台兼容）
ifeq ($(OS),Windows_NT)
    # Windows PowerShell 语法
    KRATOS_VERSION=$(shell powershell -Command "go mod graph | Select-String 'go-kratos/kratos/v2' | Select-Object -First 1 | ForEach-Object { $_.ToString().Split('@')[1] }")
else
    # macOS/Linux bash 语法
    KRATOS_VERSION=$(shell go mod graph | grep go-kratos/kratos/v2 | head -n 1 | awk -F '@' '{print $$2}')
endif

# 设置 Kratos 路径
KRATOS=$(GOPATH)/pkg/mod/github.com/go-kratos/kratos/v2@$(KRATOS_VERSION)

# 生成应用名称（跨平台兼容）
APP_NAME=$(notdir $(APP_RELATIVE_PATH))
DOCKER_IMAGE_NAME=$(subst /,-,$(APP_RELATIVE_PATH))

# 生成 Docker 镜像名称
DOCKER_IMAGE=cyi3muz/$(DOCKER_IMAGE_NAME):0.1.0

# 目标定义

.PHONY: config
# 生成 internal 目录下的 proto 文件
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# 生成 api 文档和相关代码
api:
	protoc --proto_path=${RELATIVE_PATH_TO_ROOT}api \
	       --proto_path=${RELATIVE_PATH_TO_ROOT}third_party \
	       --go_out=paths=source_relative:${RELATIVE_PATH_TO_ROOT}api \
	       --go-http_out=paths=source_relative:${RELATIVE_PATH_TO_ROOT}api \
	       --go-grpc_out=paths=source_relative:${RELATIVE_PATH_TO_ROOT}api \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)

.PHONY: errors
# 生成 errors 目录下的 proto 文件
errors:
	protoc --proto_path=${RELATIVE_PATH_TO_ROOT}api \
	       --proto_path=${RELATIVE_PATH_TO_ROOT}third_party \
	       --go_out=paths=source_relative:${RELATIVE_PATH_TO_ROOT}api \
	       --go-errors_out=paths=source_relative:${RELATIVE_PATH_TO_ROOT}api \
	       $(API_PROTO_FILES)

.PHONY: build
# 构建应用
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: docker
# 构建 Docker 镜像
docker:
	cd $(PROJECT_ROOT) && docker build -f deploy/build/Dockerfile \
		--build-arg APP_RELATIVE_PATH=$(APP_RELATIVE_PATH) \
		--build-arg APP_NAME=$(APP_NAME) \
		-t $(DOCKER_IMAGE) .

.PHONY: env
# 显示环境变量
env:
	@echo GOPATH=$(GOPATH)
	@echo VERSION=$(VERSION)
	@echo CURRENT_DIR=$(CURRENT_DIR)
	@echo PARENT_DIR=$(PARENT_DIR)
	@echo APP_RELATIVE_PATH=$(APP_RELATIVE_PATH)
	@echo APP_NAME=$(APP_NAME)
	@echo DOCKER_IMAGE=$(DOCKER_IMAGE)
	@echo PROJECT_ROOT=$(PROJECT_ROOT)
	@echo APP_MAKEFILE_PATH=$(APP_MAKEFILE_PATH)
	@echo API_PROTO_FILES=$(API_PROTO_FILES)
	@echo RELATIVE_PATH_TO_ROOT=$(RELATIVE_PATH_TO_ROOT)
# 	@ifneq ($(OS),Windows_NT)
# 	@echo Git_Bash=$(Git_Bash)
# 	@endif

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
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
