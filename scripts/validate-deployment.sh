#!/bin/bash
# ==========================================
# Deployment Validation Script
# ==========================================
# Validates that the repository is ready for deployment
# Checks:
# - Required files exist
# - Docker builds successfully
# - Migrations are valid
# - Configuration is correct
# ==========================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
CHECKS_PASSED=0
CHECKS_FAILED=0
WARNINGS=0

# Function to log with color
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
    CHECKS_PASSED=$((CHECKS_PASSED + 1))
}

log_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
    WARNINGS=$((WARNINGS + 1))
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
    CHECKS_FAILED=$((CHECKS_FAILED + 1))
}

# Check if file exists
check_file() {
    if [ -f "$1" ]; then
        log_success "File exists: $1"
        return 0
    else
        log_error "File missing: $1"
        return 1
    fi
}

# Check if directory exists
check_dir() {
    if [ -d "$1" ]; then
        log_success "Directory exists: $1"
        return 0
    else
        log_error "Directory missing: $1"
        return 1
    fi
}

# Check required files
check_required_files() {
    log_info "Checking required files..."
    echo ""
    
    check_file "Dockerfile"
    check_file "render.yaml"
    check_file "go.mod"
    check_file "go.sum"
    check_file "main.go"
    check_file "env.example"
    check_file "env.render.example"
    check_file "DEPLOYMENT.md"
    check_file "QUICKSTART.md"
    check_file ".dockerignore"
    
    echo ""
}

# Check directory structure
check_directories() {
    log_info "Checking directory structure..."
    echo ""
    
    check_dir "cmd"
    check_dir "internal"
    check_dir "internal/bootstrap"
    check_dir "internal/config"
    check_dir "migration/db"
    
    echo ""
}

# Check migration files
check_migrations() {
    log_info "Checking migration files..."
    echo ""
    
    MIGRATION_DIR="migration/db"
    
    if [ ! -d "$MIGRATION_DIR" ]; then
        log_error "Migration directory not found: $MIGRATION_DIR"
        return 1
    fi
    
    MIGRATION_COUNT=$(find "$MIGRATION_DIR" -name "*.sql" | wc -l)
    
    if [ "$MIGRATION_COUNT" -eq 0 ]; then
        log_error "No migration files found in $MIGRATION_DIR"
        return 1
    fi
    
    log_success "Found $MIGRATION_COUNT migration file(s)"
    
    # Check migration file format
    for file in "$MIGRATION_DIR"/*.sql; do
        if [ -f "$file" ]; then
            if grep -q "-- +goose Up" "$file" && grep -q "-- +goose Down" "$file"; then
                log_success "Valid goose migration: $(basename "$file")"
            else
                log_warning "Migration may be missing goose directives: $(basename "$file")"
            fi
        fi
    done
    
    echo ""
}

# Check Dockerfile
check_dockerfile() {
    log_info "Checking Dockerfile..."
    echo ""
    
    if [ ! -f "Dockerfile" ]; then
        log_error "Dockerfile not found"
        return 1
    fi
    
    # Check for multi-stage build
    if grep -q "FROM.*AS builder" Dockerfile; then
        log_success "Multi-stage build detected"
    else
        log_warning "Multi-stage build not detected"
    fi
    
    # Check for health check
    if grep -q "HEALTHCHECK" Dockerfile; then
        log_success "Health check configured"
    else
        log_warning "Health check not configured"
    fi
    
    # Check for entrypoint
    if grep -q "ENTRYPOINT" Dockerfile; then
        log_success "Entrypoint configured"
    else
        log_warning "Entrypoint not configured"
    fi
    
    # Check for migration copy
    if grep -q "COPY.*migration" Dockerfile; then
        log_success "Migration files will be copied to container"
    else
        log_error "Migration files not copied to container"
    fi
    
    echo ""
}

# Check render.yaml
check_render_yaml() {
    log_info "Checking render.yaml..."
    echo ""
    
    if [ ! -f "render.yaml" ]; then
        log_error "render.yaml not found"
        return 1
    fi
    
    # Check for required fields
    if grep -q "type: web" render.yaml; then
        log_success "Web service type configured"
    else
        log_error "Web service type not configured"
    fi
    
    if grep -q "env: docker" render.yaml; then
        log_success "Docker environment configured"
    else
        log_error "Docker environment not configured"
    fi
    
    if grep -q "healthCheckPath:" render.yaml; then
        log_success "Health check path configured"
    else
        log_warning "Health check path not configured"
    fi
    
    echo ""
}

# Check Go dependencies
check_go_dependencies() {
    log_info "Checking Go dependencies..."
    echo ""
    
    if ! command -v go &> /dev/null; then
        log_warning "Go not installed (OK if deploying with Docker)"
        return 0
    fi
    
    if go mod verify; then
        log_success "Go modules verified"
    else
        log_error "Go modules verification failed"
    fi
    
    echo ""
}

# Check Docker (optional)
check_docker() {
    log_info "Checking Docker..."
    echo ""
    
    if ! command -v docker &> /dev/null; then
        log_warning "Docker not installed (optional for local testing)"
        return 0
    fi
    
    log_success "Docker is installed"
    
    # Try to build Docker image
    log_info "Attempting to build Docker image (this may take a few minutes)..."
    if docker build -t journal-api-test . > /tmp/docker_build.log 2>&1; then
        log_success "Docker image built successfully"
        # Clean up test image
        docker rmi journal-api-test > /dev/null 2>&1 || true
    else
        log_error "Docker build failed (see /tmp/docker_build.log for details)"
        tail -20 /tmp/docker_build.log
    fi
    
    echo ""
}

# Check environment example files
check_env_examples() {
    log_info "Checking environment example files..."
    echo ""
    
    # Check env.example
    if [ -f "env.example" ]; then
        log_success "env.example exists"
        
        # Check for required variables
        for var in DATABASE_DSN TOKEN_PASSWORD_SALT TOKEN_ACCESS_TOKEN_SECRET TOKEN_REFRESH_TOKEN_SECRET; do
            if grep -q "$var" env.example; then
                log_success "env.example contains $var"
            else
                log_warning "env.example missing $var"
            fi
        done
    else
        log_error "env.example not found"
    fi
    
    echo ""
    
    # Check env.render.example
    if [ -f "env.render.example" ]; then
        log_success "env.render.example exists"
        
        # Check for Render-specific variables
        if grep -q "DATABASE_DIRECT_DSN" env.render.example; then
            log_success "env.render.example contains DATABASE_DIRECT_DSN"
        else
            log_warning "env.render.example missing DATABASE_DIRECT_DSN"
        fi
    else
        log_warning "env.render.example not found"
    fi
    
    echo ""
}

# Check documentation
check_documentation() {
    log_info "Checking documentation..."
    echo ""
    
    check_file "README.md"
    check_file "DEPLOYMENT.md"
    check_file "QUICKSTART.md"
    
    # Check if README mentions deployment
    if grep -q -i "deployment" README.md; then
        log_success "README.md mentions deployment"
    else
        log_warning "README.md doesn't mention deployment"
    fi
    
    echo ""
}

# Print summary
print_summary() {
    echo ""
    echo "=========================================="
    echo "Validation Summary"
    echo "=========================================="
    echo ""
    echo -e "${GREEN}Checks Passed:${NC} $CHECKS_PASSED"
    echo -e "${YELLOW}Warnings:${NC} $WARNINGS"
    echo -e "${RED}Checks Failed:${NC} $CHECKS_FAILED"
    echo ""
    
    if [ $CHECKS_FAILED -eq 0 ]; then
        echo -e "${GREEN}✅ Repository is ready for deployment!${NC}"
        echo ""
        echo "Next steps:"
        echo "1. Run 'make gen-secrets' to generate secure tokens"
        echo "2. Set environment variables in Render dashboard"
        echo "3. Push to GitHub and deploy via Render"
        echo "4. See QUICKSTART.md for detailed instructions"
        echo ""
        return 0
    else
        echo -e "${RED}❌ Repository has issues that need to be fixed${NC}"
        echo ""
        echo "Please fix the errors above before deploying."
        echo ""
        return 1
    fi
}

# Main validation
main() {
    echo ""
    echo "=========================================="
    echo "Deployment Validation Script"
    echo "=========================================="
    echo ""
    
    check_required_files
    check_directories
    check_migrations
    check_dockerfile
    check_render_yaml
    check_go_dependencies
    check_env_examples
    check_documentation
    
    # Optional Docker check (don't fail if Docker not available)
    if command -v docker &> /dev/null; then
        check_docker
    fi
    
    print_summary
}

# Run main function
main "$@"

