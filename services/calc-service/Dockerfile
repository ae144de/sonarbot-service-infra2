# Builder
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o calc-service cmd/main.go

# Final
FROM alpine:3.17
WORKDIR /app
COPY --from=builder /app/calc-service ./calc-service
EXPOSE 8080
ENTRYPOINT ["./calc-service"]