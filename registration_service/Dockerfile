# Start from the official Go image
FROM golang:1.24-alpine

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o main .

# Expose the port
EXPOSE 8080

# Run the app
CMD ["./main"]
