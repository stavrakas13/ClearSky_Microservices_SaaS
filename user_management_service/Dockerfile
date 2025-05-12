# Stage 1: builder με CGO_ENABLED=1
FROM golang:1.24-alpine AS builder

# Ενεργοποιούμε CGO και εγκαθιστούμε τα C-dependencies
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev sqlite-dev git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Χτίζουμε το binary (θα συμπεριλάβει και το sqlite-driver)
RUN go build -o user_management_service ./cmd/server

# Stage 2: runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /root/
# Αντιγράφουμε το binary
COPY --from=builder /app/user_management_service .

# Δηλώνουμε τον φάκελο όπου θα γραφτεί η SQLite DB (προαιρετικό)
VOLUME ["/root"]

# Άνοιγμα port
EXPOSE 8082

# Εκκίνηση της εφαρμογής
CMD ["./user_management_service"]
