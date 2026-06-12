# Build Stage
FROM golang:alpine AS builder

# Install make
RUN apk add --no-cache make

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN make build-linux

# Final Stage (Minimal Image)
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary
COPY --from=builder /app/bin/linux-amd64/secnotes-server .

# Create the data directory for persistent storage
RUN mkdir -p /data/server-data

# Expose the default Fly HTTP port
EXPOSE 8080

# Run the server, mapping the store to our persistent volume
CMD ["./secnotes-server", "--port", "8080", "--store", "/data/server-data"]
