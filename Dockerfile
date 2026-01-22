# 多阶段构建 - 构建阶段
FROM golang:1.24.2-alpine AS builder

# 设置工作目录
WORKDIR /app

# 设置Go模块代理为国内源
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 复制源代码
COPY . .

# 设置国内镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
 	apk add --no-cache git ca-certificates tzdata && \
	go mod download && \
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# 运行阶段
FROM alpine:3.20.2

# 设置时区
ENV TZ=Asia/Shanghai

# 安装ca-certificates和tzdata
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
# COPY --from=builder /app/sql ./sql

# 暴露端口
EXPOSE 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# 启动命令
CMD ["./main"]
