# 构建阶段
FROM golang:1.21-alpine AS builder

# 安装基础工具
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /build

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/main.go

# 运行阶段
FROM alpine:latest

# 安装基础工具和时区
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1000 -S adplatform && \
    adduser -u 1000 -S adplatform -G adplatform

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/main .

# 复制配置文件
COPY --from=builder /build/configs ./configs

# 创建必要的目录
RUN mkdir -p uploads logs && \
    chown -R adplatform:adplatform /app

# 切换到非root用户
USER adplatform

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# 启动应用
CMD ["./main"]