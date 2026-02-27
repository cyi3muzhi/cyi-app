FROM golang:1.25 AS builder

ARG APP_RELATIVE_PATH

COPY . /src
WORKDIR /src/app/${APP_RELATIVE_PATH}

RUN GOPROXY=https://goproxy.cn make build

FROM debian:stable-slim

ARG APP_RELATIVE_PATH
ARG APP_NAME

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

# 设置时区为上海
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
RUN echo "Asia/Shanghai" > /etc/timezone

COPY --from=builder /src/app/${APP_RELATIVE_PATH}/bin /app

WORKDIR /app

# 设置环境变量
ENV APP_NAME=${APP_NAME}

# 给可执行文件添加执行权限
RUN chmod +x ./${APP_NAME}

# 启动命令（使用 shell 格式，环境变量会被展开）
CMD ./${APP_NAME}
