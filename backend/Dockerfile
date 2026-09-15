FROM golang:1.22-alpine AS builder
ENV GOPROXY=http://mirrors.aliyun.com/goproxy,direct
ENV GOINSECURE=mirrors.aliyun.com
ENV GOSUMDB=off
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["/app/server"]
