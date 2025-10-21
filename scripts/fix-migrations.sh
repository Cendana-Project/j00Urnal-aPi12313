#!/bin/bash
set -e

echo "🔧 Database Migration Fix Script"
echo "================================"

# Function to check if we're running in Docker
is_docker() {
    [ -f /.dockerenv ] || grep -q docker /proc/1/cgroup 2>/dev/null
}

# Function to connect to database and check migration state
check_migration_state() {
    echo "📊 Checking current migration state..."
    
    # Get database connection details from environment or config
    DB_DSN="${DATABASE_DSN:-postgres://user:password@localhost:5432/api_monolith?sslmode=disable}"
    
    # Check if goose_schema_version table exists and has the problematic migration
    psql "$DB_DSN" -c "
        SELECT version_id, is_applied, tstamp 
        FROM goose_db_version 
        WHERE version_id = 20250616162131;
    " 2>/dev/null || echo "Migration 20250616162131 not found in database"
}

# Function to clean up problematic migration
cleanup_migration() {
    echo "🧹 Cleaning up problematic migration..."
    
    DB_DSN="${DATABASE_DSN:-postgres://user:password@localhost:5432/api_monolith?sslmode=disable}"
    
    # Remove the problematic migration record if it exists
    psql "$DB_DSN" -c "
        DELETE FROM goose_db_version 
        WHERE version_id = 20250616162131;
    " 2>/dev/null && echo "✅ Removed problematic migration record" || echo "ℹ️  No problematic migration record found"
}

# Function to run migrations safely
run_migrations_safely() {
    echo "🚀 Running migrations safely..."
    
    # First, try to run migrations normally
    if go run main.go migrate --action up; then
        echo "✅ Migrations completed successfully"
        return 0
    else
        echo "⚠️  Migration failed, attempting cleanup..."
        cleanup_migration
        
        # Try again after cleanup
        if go run main.go migrate --action up; then
            echo "✅ Migrations completed successfully after cleanup"
            return 0
        else
            echo "❌ Migration still failed after cleanup"
            return 1
        fi
    fi
}

# Main execution
main() {
    echo "Starting migration fix process..."
    
    # Check if we're in Docker
    if is_docker; then
        echo "🐳 Running in Docker container"
    else
        echo "💻 Running on host system"
    fi
    
    # Check migration state
    check_migration_state
    
    # Run migrations safely
    if run_migrations_safely; then
        echo "🎉 Migration fix completed successfully!"
        exit 0
    else
        echo "💥 Migration fix failed"
        exit 1
    fi
}

# Run main function
main "$@" 