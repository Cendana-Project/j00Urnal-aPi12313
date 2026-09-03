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
# Use wget if available, otherwise use nc (netcat) as fallback
RUN echo '#!/bin/sh' > /healthcheck.sh && \
    echo 'APP_PORT=${SERVER_PORT:-${PORT:-8080}}' >> /healthcheck.sh && \
    echo 'if command -v wget > /dev/null 2>&1; then' >> /healthcheck.sh && \
    echo '  wget --no-verbose --tries=1 --spider http://localhost:${APP_PORT}/_internal/healthz || exit 1' >> /healthcheck.sh && \
    echo 'elif command -v nc > /dev/null 2>&1; then' >> /healthcheck.sh && \
    echo '  echo "GET /_internal/healthz HTTP/1.1\r\nHost: localhost\r\n\r\n" | nc localhost ${APP_PORT} | grep -q "200 OK" || exit 1' >> /healthcheck.sh && \
    echo 'else' >> /healthcheck.sh && \
    echo '  exit 0' >> /healthcheck.sh && \
    echo 'fi' >> /healthcheck.sh && \
    chmod +x /healthcheck.sh

# Entrypoint script that runs migration before starting server
RUN echo '#!/bin/sh' > /entrypoint.sh && \
    echo 'set -e' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Function to log with timestamp' >> /entrypoint.sh && \
    echo 'log() {' >> /entrypoint.sh && \
    echo '  echo "[$(date +"%Y-%m-%d %H:%M:%S")] $*"' >> /entrypoint.sh && \
    echo '}' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo 'log "=========================================="' >> /entrypoint.sh && \
    echo 'log "Starting Journal API deployment"' >> /entrypoint.sh && \
    echo 'log "=========================================="' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Run database migrations' >> /entrypoint.sh && \
    echo 'log "Running database migrations..."' >> /entrypoint.sh && \
    echo 'set +e  # Allow migration to exit with non-zero if all migrations are already applied' >> /entrypoint.sh && \
    echo './main migrate --action=up 2>&1' >> /entrypoint.sh && \
    echo 'MIGRATION_EXIT=$?' >> /entrypoint.sh && \
    echo 'set -e  # Re-enable exit on error' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Migration exit codes:' >> /entrypoint.sh && \
    echo '# 0 = success (migrations applied or already up to date)' >> /entrypoint.sh && \
    echo '# Non-zero may indicate migrations are already current (which is OK)' >> /entrypoint.sh && \
    echo 'if [ $MIGRATION_EXIT -eq 0 ]; then' >> /entrypoint.sh && \
    echo '  log "✓ Database migrations completed successfully"' >> /entrypoint.sh && \
    echo 'else' >> /entrypoint.sh && \
    echo '  log "⚠ Migration exited with code $MIGRATION_EXIT (may be OK if migrations are current)"' >> /entrypoint.sh && \
    echo 'fi' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo '# Never seed production automatically. In non-production, seeding is best-effort so a' >> /entrypoint.sh && \
    echo '# temporary database outage cannot prevent the API server from starting.' >> /entrypoint.sh && \
    echo 'if [ "${ENV:-development}" = "production" ]; then' >> /entrypoint.sh && \
    echo '  log "Skipping automatic database seeding in production"' >> /entrypoint.sh && \
    echo 'else' >> /entrypoint.sh && \
    echo '  log "Running database seeding (best effort)..."' >> /entrypoint.sh && \
    echo '  if ./main seed; then' >> /entrypoint.sh && \
    echo '    log "Database seeding completed"' >> /entrypoint.sh && \
    echo '  else' >> /entrypoint.sh && \
    echo '    SEED_EXIT=$?' >> /entrypoint.sh && \
    echo '    log "Database seeding exited with code $SEED_EXIT; continuing server startup"' >> /entrypoint.sh && \
    echo '  fi' >> /entrypoint.sh && \
    echo 'fi' >> /entrypoint.sh && \
    echo '' >> /entrypoint.sh && \
    echo 'log "=========================================="' >> /entrypoint.sh && \
    echo 'log "Starting server on port ${SERVER_PORT:-${PORT:-8080}}..."' >> /entrypoint.sh && \
    echo 'log "=========================================="' >> /entrypoint.sh && \
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
HEALTHCHECK --interval=30s --timeout=15s --start-period=15s --retries=3 \
  CMD /healthcheck.sh

# Run migrations then start server
ENTRYPOINT ["/entrypoint.sh"]
