FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o kline-fetcher cmd/main.go

FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/kline-fetcher ./kline-fetcher
ENTRYPOINT ["./kline-fetcher"]