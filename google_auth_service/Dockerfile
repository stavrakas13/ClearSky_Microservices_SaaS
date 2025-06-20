# 1η Φάση: Build
FROM golang:1.24-alpine AS builder

# Enable CGO and install C dependencies for SQLite
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev sqlite-dev git

WORKDIR /app

# Αντιγραφή μόνο των αναγκαίων αρχείων
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary (with sqlite driver)
RUN go build -o google_auth_service main.go

# 2η Φάση: Τελικό Image
FROM alpine:latest

WORKDIR /root/

# Install SQLite for runtime
RUN apk --no-cache add ca-certificates sqlite

# Αντιγραφή μόνο του binary
COPY --from=builder /app/google_auth_service .

# Ορίζουμε Port
EXPOSE 8086

# Default εντολή για να τρέχει το binary
CMD ["./google_auth_service"]

