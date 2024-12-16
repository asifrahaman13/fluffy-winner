# Use the official Golang image as the base image
FROM golang:1.17

# Set the working directory inside the container
WORKDIR /app

# Copy the Go modules and the Go files
COPY go.mod go.sum ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY test/ ./test/

# Build the Go application
RUN go mod download
RUN go build -o myapp cmd/main.go

# Expose the port the application runs on
EXPOSE 8080

# Command to run the application
CMD ["./myapp"]
