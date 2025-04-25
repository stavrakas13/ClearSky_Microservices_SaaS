# Χρησιμοποιούμε επίσημο Go image
FROM golang:1.22

# Δημιουργούμε φάκελο εργασίας μέσα στο container
WORKDIR /app

# Αντιγράφουμε τα αρχεία Go Modules πρώτα (optimization για caching)
COPY go.mod go.sum ./
RUN go mod download

# Μετά αντιγράφουμε όλο το project
COPY . .

# Κάνουμε build το project
RUN go build -o user_management_service ./cmd/server

# Το container ακούει στο port 8082
EXPOSE 8082

# Default command όταν ξεκινάει το container
CMD ["./user_management_service"]
