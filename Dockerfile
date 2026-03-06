# =======================
# 1. 构建阶段
# =======================
FROM docker.m.daocloud.io/library/golang:1.25-alpine AS builder

WORKDIR /build
# 设置 Go 代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

COPY . .
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache git \
    && go mod download \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o devops ./cmd/server


FROM docker.m.daocloud.io/library/alpine:3.20.2

ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /build/devops .
COPY config/ ./config/
# 使用阿里 Alpine 源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache ca-certificates tzdata wget \
    && chmod +x ./devops

EXPOSE 8081

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

CMD ["./devops"]