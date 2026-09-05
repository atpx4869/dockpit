# ============ 构建阶段 ============
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git gcc musl-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 确保 front 目录存在（go:embed 需要）
RUN mkdir -p front && touch front/.gitkeep

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/dockpit .

# ============ 运行阶段 ============
FROM alpine:3.20

LABEL authors="atpx4869"

RUN apk add --no-cache \
    tzdata \
    docker-cli \
    docker-cli-compose \
    ca-certificates

WORKDIR /app

COPY --from=builder /app/dockpit ./dockpit
COPY etc/ ./etc/
COPY start.sh ./start.sh

RUN chmod +x start.sh dockpit

ENV secretKey="" \
    DOCKER_HOST="unix:///var/run/docker.sock" \
    BACKUP_DIR="/data/backups" \
    TZ="Asia/Shanghai"

VOLUME ["/data"]

EXPOSE 12712

CMD ["./start.sh"]
