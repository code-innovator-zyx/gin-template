# 多阶段构建 Dockerfile - Scratch 版本（最小镜像）
# 阶段1: 编译
FROM golang:1.24-alpine AS builder

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

ADD go.mod .
ADD go.sum .
RUN go mod download

COPY . .
COPY ./config /app/config

RUN go build \
    -a \
    -installsuffix cgo \
    -ldflags="-s -w -extldflags '-static'" \
    -o /app/admin \
    main.go

FROM scratch

WORKDIR /app


ENV TZ=Asia/Shanghai

COPY --from=builder /app/admin /app/admin
COPY --from=builder /app/config /app/config

# 启动应用
ENTRYPOINT ["/app/admin"]
