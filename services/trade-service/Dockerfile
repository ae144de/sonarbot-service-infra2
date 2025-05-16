# Builder
FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o trade-service cmd/main.go

# Final
FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/trade-service ./trade-service
EXPOSE 8080
ENTRYPOINT ["./trade-service"]