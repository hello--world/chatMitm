# 构建阶段
FROM golang:1.22-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的工具
RUN apk add --no-cache git

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o chatMitm main.go

# 运行阶段
FROM alpine:latest

# 安装 CA 证书（用于 HTTPS 连接）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/chatMitm /app/chatMitm

# 创建数据目录和证书目录
RUN mkdir -p /app/stream_data /home/appuser/.mitmproxy && \
    chown -R appuser:appuser /app /home/appuser

# 切换到非 root 用户
USER appuser

# 暴露端口
EXPOSE 9080 9081

# 设置环境变量
ENV STREAM_DATA_DIR=/app/stream_data

# 启动应用
CMD ["/app/chatMitm"]

