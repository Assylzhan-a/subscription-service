FROM golang:1.23.1-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o subscription-service ./cmd/api

# Use a smaller image for the final build
FROM alpine:3.18

WORKDIR /app

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/subscription-service .

# Expose port
EXPOSE 8080

# Command to run the executable
CMD ["./subscription-service"]