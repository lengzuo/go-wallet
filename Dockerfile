# Build Stage
FROM golang:1.24 as builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Production
ENV MODE=prod

# Build the application
# CGO_ENABLED=0 is important for creating a statically linked binary
# -o /app/main specifies the output path and name of the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./cmd/main.go

# Final Stage
# Use a minimal image like alpine or scratch
FROM alpine:latest

# Install ca-certificates for HTTPS connections (if needed)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Expose the port your application listens on (e.g., 8080)
EXPOSE 8080

# Command to run the application
CMD ["./main"]
