FROM docker.m.daocloud.io/library/alpine:3.20.2

# 设置时区
ENV TZ=Asia/Shanghai

# 安装ca-certificates和tzdata
RUN apk --no-cache add ca-certificates tzdata

# 设置工作目录
WORKDIR /app

COPY devops .
COPY configs/ ./configs/

# 暴露端口
EXPOSE 8081
RUN chmod +x ./devops
# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/health || exit 1

# 启动命令
CMD ["./devops"]
