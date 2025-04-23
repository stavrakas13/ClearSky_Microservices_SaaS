# builder
FROM golang:1.24-alpine3.20 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# final
FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/main .
# (You could copy the .env if you want, but Compose injects it for you.)
EXPOSE 8081
CMD ["./main"]
