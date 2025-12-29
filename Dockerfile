# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates and wget for HTTPS requests and health checks
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Health check script (create before switching users)
RUN echo '#!/bin/sh' > /healthcheck.sh && \
    echo 'PORT=${PORT:-8080}' >> /healthcheck.sh && \
    echo 'wget --no-verbose --tries=1 --spider http://localhost:${PORT}/_internal/healthz || exit 1' >> /healthcheck.sh && \
    chmod +x /healthcheck.sh

# Entrypoint script that runs migration before starting server
RUN echo '#!/bin/sh' > /entrypoint.sh && \
    echo 'set +e' >> /entrypoint.sh && \
    echo 'echo "Running database migrations..."' >> /entrypoint.sh && \
    echo './main migrate --action=up' >> /entrypoint.sh && \
    echo 'MIGRATION_EXIT=$?' >> /entrypoint.sh && \
    echo 'if [ $MIGRATION_EXIT -ne 0 ]; then' >> /entrypoint.sh && \
    echo '  echo "Migration completed with exit code $MIGRATION_EXIT (some migrations may have been skipped)"' >> /entrypoint.sh && \
    echo 'fi' >> /entrypoint.sh && \
    echo 'echo "Starting server..."' >> /entrypoint.sh && \
    echo 'exec ./main server' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migration files (needed for goose migrations at runtime)
COPY --from=builder /app/migration ./migration

# Change ownership to appuser
RUN chown -R appuser:appuser /root/

# Switch to non-root user
USER appuser

# Expose port (Render will set PORT env var)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD /healthcheck.sh

# Run migrations then start server
ENTRYPOINT ["/entrypoint.sh"]
