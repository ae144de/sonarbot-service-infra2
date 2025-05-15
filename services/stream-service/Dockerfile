  # builder
  FROM golang:1.19-alpine AS builder
  WORKDIR /app
  COPY go.mod go.sum ./
  RUN go mod download
  COPY . .
  RUN go build -o stream-service cmd/main.go
  
  # final image
  FROM alpine:3.17
  WORKDIR /app
  COPY --from=builder /app/stream-service ./stream-service
  EXPOSE 8080
  ENTRYPOINT ["./stream-service"]