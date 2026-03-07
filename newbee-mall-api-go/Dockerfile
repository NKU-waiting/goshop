# ---- Build Stage ----
# 使用官方的 Go 镜像作为构建环境
FROM golang:1.20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 复制 vendor 目录
COPY vendor ./

# 复制所有源代码到工作目录
COPY . .

# 构建 Go 应用，使用 -mod=vendor 标志
# 这个标志强制 Go 编译器只使用 vendor 目录中的依赖，完全避免网络访问
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /newbee-mall-api-go -ldflags "-s -w" .

# ---- Final Stage ----
# 使用一个非常小的基础镜像，如 alpine
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /newbee-mall-api-go .

# 暴露应用程序端口
EXPOSE 8888

# 容器启动时运行的命令
CMD ["./newbee-mall-api-go"]
