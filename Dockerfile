# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags="-s -w" \
    -o SwagFluence ./cmd/swagfluence

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/swagfluence .

# Set environment variables (can be overridden at runtime)
ENV CONFLUENCE_BASE_URL=""
ENV CONFLUENCE_USERNAME=""
ENV CONFLUENCE_API_TOKEN=""
ENV CONFLUENCE_SPACE_KEY=""
ENV CONFLUENCE_PARENT_PAGE_ID=""

ENTRYPOINT ["./SwagFluence"]