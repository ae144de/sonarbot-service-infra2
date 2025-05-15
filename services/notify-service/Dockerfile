# Builder
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o notify-service cmd/main.go

# Final
FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/notify-service ./notify-service
EXPOSE 8080
ENTRYPOINT ["./notify-service"]