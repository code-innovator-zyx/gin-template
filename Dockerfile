# 多阶段构建 Dockerfile
# 阶段1: 编译
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git make

# 复制依赖文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/gin-template .

# 阶段2: 运行
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

# 设置工作目录
WORKDIR /app

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/gin-template .

# 复制配置模板
COPY app.yaml.template ./app.yaml.template

# 创建日志目录
RUN mkdir -p logs && chown -R app:app /app

# 切换到非root用户
USER app

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# 启动应用
CMD ["./gin-template"]

