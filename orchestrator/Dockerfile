# Stage 1: Build
FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o orchestrator cmd/orchestrator/main.go

# Stage 2: Runtime
FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/orchestrator .
ENV RABBITMQ_URL=amqp://admin:admin@rabbitmq:5672/
ENTRYPOINT ["./orchestrator"]