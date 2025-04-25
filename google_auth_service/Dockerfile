# 1η Φάση: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Αντιγραφή μόνο των αναγκαίων αρχείων
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Κτίζουμε το binary
RUN go build -o google_auth_service main.go

# 2η Φάση: Τελικό Image
FROM alpine:latest

WORKDIR /root/

# Αντιγραφή μόνο του binary
COPY --from=builder /app/google_auth_service .

# Ορίζουμε Port
EXPOSE 8086

# Default εντολή για να τρέχει το binary
CMD ["./google_auth_service"]

