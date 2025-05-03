FROM golang:1.23-alpine

# Δημιουργία working directory
WORKDIR /app

# Αντιγραφή των go.mod και go.sum πρώτα για caching
COPY go.mod go.sum ./
RUN go mod download

# Αντιγραφή του υπόλοιπου κώδικα
COPY . .

# Compile της εφαρμογής
RUN go build -o app

# Άνοιγμα του port
EXPOSE 8080

# Εκκίνηση της εφαρμογής
CMD ["./app"]
