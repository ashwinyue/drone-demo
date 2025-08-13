# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-drone-deploy ./cmd/go-drone-deploy

# 运行阶段
FROM alpine:latest

# 安装必要的工具
RUN apk --no-cache add ca-certificates docker kubectl

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/go-drone-deploy .
COPY --from=builder /app/configs ./configs

# 运行应用
CMD ["./go-drone-deploy"]
