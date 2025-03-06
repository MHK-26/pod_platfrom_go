# deployments/docker/content-service/Dockerfile
FROM golang:1.20-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o content-service ./cmd/content-service/main.go

# Start a new stage from scratch
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/content-service .
COPY --from=builder /app/.env .

# Create storage directories
RUN mkdir -p /app/storage/podcasts /app/storage/episodes

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./content-service"]