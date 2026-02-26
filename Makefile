GOHOSTOS:=$(shell go env GOHOSTOS) # 判断当前系统是否为 windows
GOPATH:=$(shell go env GOPATH)  # 获取 GOPATH 环境变量
VERSION=$(shell git describe --tags --always) # 获取当前 Git 仓库的版本标签
SERVICE_CMD_DIRS := $(wildcard app/*/*/cmd/* app/*/cmd/*) # 查找所有服务的 cmd 目录
SERVICE_NAMES := $(foreach dir,$(SERVICE_CMD_DIRS),$(notdir $(dir))) # 从 cmd 目录中提取服务名称
$(foreach service,$(SERVICE_NAMES),$(eval wire-$(service):; cd $(filter %/cmd/$(service),$(SERVICE_CMD_DIRS)) && wire)) # 为每个服务生成 wire 代码
$(foreach service,$(SERVICE_NAMES),$(eval docker-build-$(service):; cd $(patsubst %/cmd/$(service),%,$(filter %/cmd/$(service),$(SERVICE_CMD_DIRS))) && make docker)) # 为每个服务生成 docker-build 目标
APP_RELATIVE_PATH=$(notdir $(patsubst %/,%,$(dir $(CURDIR))))/$(notdir $(CURDIR))
ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git)))) 
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")  
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")  
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)  
	API_PROTO_FILES=$(shell find api -name *.proto)  
endif

.PHONY: init
# init env 初始化环境变量
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@1.28.1
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@1.28.1
	go install github.com/go-kratos/kratos/cmd/kratos/v2@2.8.3
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@1.28.1
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
# generate internal proto 生成 internal 目录下的 proto 文件
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
   	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto 生成 api 目录下的 proto 文件
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
   	       --go_out=paths=source_relative:./api \
   	       --go-http_out=paths=source_relative:./api \
   	       --go-grpc_out=paths=source_relative:./api \
	       --openapi_out=fq_schema_naming=true,default_response=false:. \
	       $(API_PROTO_FILES)


.PHONY: validate 
# generate validate proto 生成 validate 目录下的 proto 文件
validate:
	protoc --proto_path=. \
			--proto_path=./third_party \
			--go_out=paths=source_relative:. \
			--validate_out=paths=source_relative,lang=go:. \
			$(API_PROTO_FILES)


.PHONY: errors
# generate errors proto 生成 errors 目录下的 proto 文件
errors:
	protoc --proto_path=. \
         --proto_path=./third_party \
         --go_out=paths=source_relative:. \
         --go-errors_out=paths=source_relative:. \
         $(API_PROTO_FILES)

.PHONY: build
# build 构建项目
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

.PHONY: generate
# generate 生成项目代码
generate:
	go generate ./...
	go mod tidy

.PHONY: all
# generate all 生成所有 proto 文件
all:
	make api;
	make config;
	make generate;

.PHONY: service
# generate service template 生成服务模板
service:
	@powershell -Command "$$goals = '$(MAKECMDGOALS)'.Split(' '); $$args = $$goals | Where-Object { $$_ -ne 'service' }; if ($$args.Count -eq 0) { echo 'Error: Service name is required. Usage: make service <service-name> [options]'; echo 'Options: --description=<description> --version=<version> --author=<author> --features=<features>'; exit 1; } cd tools/service-gen; go run main.go $$args"


.PHONY: wire
# generate wire code 生成 wire 代码
wire:
	$(foreach service,$(SERVICE_NAMES),wire-$(service))

.PHONY: docker-up
# run docker container 运行 docker 容器
docker-up:
	cd deploy && docker-compose -p cyi-app up -d

.PHONY: docker-down
# stop docker container 停止 docker 容器
docker-down:
	cd deploy && docker-compose -p cyi-app down

.PHONY: docker-build
# build docker image 构建 docker 镜像
docker-build:
	$(foreach service,$(SERVICE_NAMES),$(MAKE) docker-build-$(service))

# show env 显示环境变量
env:
	@echo 'GOHOSTOS: $(GOHOSTOS)'
	@echo 'GOPATH: $(GOPATH)'
	@echo 'VERSION: $(VERSION)'
	@echo 'APP_RELATIVE_PATH: $(APP_RELATIVE_PATH)'
# 	测试相对路径是否正确
	@echo 'SERVICE_CMD_DIRS: $(SERVICE_CMD_DIRS)'
	@echo 'SERVICE_NAMES: $(SERVICE_NAMES)'

# show help
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

.DEFAULT_GOAL := help

# Ignore any arguments passed to service target
%:
	@: