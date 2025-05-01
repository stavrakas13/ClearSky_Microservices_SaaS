# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Εγκατάσταση build dependencies (για alpine)
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binary
RUN go build -o user_management_service ./cmd/server

# Stage 2: Runtime
FROM alpine:latest

# Εγκατάσταση runtime βιβλιοθηκών
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Αντιγραφή binary από builder
COPY --from=builder /app/user_management_service .

# Άνοιγμα port (matching με docker-compose)
EXPOSE 8082

# Εκκίνηση
CMD ["./user_management_service"]
