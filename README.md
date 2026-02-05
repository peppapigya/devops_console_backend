#  一、 Linux部署

## 1. 克隆仓库

首先要准备go环境，参考地址https://go-lang.org.cn/doc/install

```bash
# 克隆仓库地址
git clone https://gitee.com/dong-xianguo/devops_backend.git
# 进入仓库地址
cd devops_backend/
```

## 2. 导入数据库文件

数据库文件在 **sql** 目录下

## 3. 修改配置信息

配置文件目录：**config/config.yaml**

主要是修改**数据库和redis**

```bash
vim config/config.yaml
server:
  port: ":8081" # 服务开放的端口，自行修改
  log_level: "info"  # debug, info, warn, error

# 数据库配置
database:
  type: "mysql"  # 仅支持MySQL
  auto_migrate: true  # 是否自动迁移数据库模型
  # MySQL配置
  mysql:
    host: "47.104.247.159"  # Docker容器中的服务名
    port: 8002
    username: "root"  # Docker环境中的用户
    password: "peppa-pig"  # Docker环境中的密码
    database: "devops_console"
    charset: "utf8mb4"
    parse_time: true
    max_open_conns: 10
    max_idle_conns: 5

# 日志配置
logging:
  format: "json"  # json, text
  time_format: "2006-01-02 15:04:05"
  report_caller: true

# 应用配置
app:
  name: "DevOps Console"
  version: "1.0.0"
  environment: "development"  # development, production

# Elasticsearch配置
elasticsearch:
  timeout: 30  # 连接超时时间（秒）
  retry: 3    # 重试次数
  health_check_interval: 60  # 健康检查间隔（秒）

# Kubernetes配置
kubernetes:
  config_path: ""  # kubeconfig文件路径，为空时使用集群内配置
  timeout: 30      # 操作超时时间（秒）
  retry: 3         # 重试次数

# Swagger配置
swagger:
  enabled: true
  host: "localhost:8081"
  base_path: "/"

# 健康检查配置
health:
  enabled: true
  endpoint: "/health"
  interval: 30  # 检查间隔（秒）

redis:
  host: 127.0.0.1
  port: 6379
  password: "123456"
  db: 0

jwt:
  secret: "n02y2Zqf4eL0hZ4xjQH9w1zDk1w5FqMnc9R+N8T1v2E="
  expire-time: 3600 # 单位为秒
  refresh-expire-time: 604800
  exclude-paths:
    - /api/v1/system/login
    - /api/v1/sysUser/refresh
    - /api/v1/sysUser/captcha
    - /swagger/*
    - /jobs/script/
    - /metrics
```

## 4. 编译

```bash
# 添加加速源
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct
# 下载对应的依赖包
go mod tidy
go mod download

# 编译,编译之后会生成一个名字为devops的二进制包，这个名字可以自定义
go build -ldflags="-s -w" -o devops ./cmd/server/main.go
```

## 5. 运行

```bash
# 直接运行二进制
./devops
```
用户名: admin，密码：admin123


