FROM golang:1.25-alpine AS builder

ARG APP_RELATIVE_PATH

RUN apk add --no-cache make git

# 先复制依赖清单，利用 Docker 层缓存加速重复构建
COPY go.mod go.sum* ./
RUN go mod download 2>/dev/null || true

COPY . /src
WORKDIR /src/app/${APP_RELATIVE_PATH}

RUN CGO_ENABLED=0 GOPROXY=off make build

FROM alpine:3.20

ARG APP_RELATIVE_PATH
ARG APP_NAME

RUN apk add --no-cache ca-certificates tzdata

# 设置时区为上海
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY --from=builder /src/app/${APP_RELATIVE_PATH}/bin /app

WORKDIR /app

# 设置环境变量
ENV APP_NAME=${APP_NAME}

# 给可执行文件添加执行权限
RUN chmod +x ./${APP_NAME}

# 启动命令（使用 shell 格式，环境变量会被展开）
CMD ./${APP_NAME}
