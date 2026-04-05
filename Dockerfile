# Base Stage
FROM golang:1.22-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Development Stage (Using Air for Live Reloading)
FROM base AS dev
RUN go install github.com/air-verse/air@latest
COPY . .
# Expose port and run Air
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# Builder Stage (compiles binary)
FROM base AS builder
COPY . .
# Build the application statically
RUN CGO_ENABLED=0 GOOS=linux go build -o /api main.go

# Production Stage
FROM alpine:latest AS prod
WORKDIR /app
# Copy the compiled binary from the builder stage
COPY --from=builder /api /api
# Expose port 8080 to the outside world
EXPOSE 8080
# Command to run the executable
CMD ["/api"]
