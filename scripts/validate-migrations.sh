#!/bin/bash
set -e

echo "🔍 Migration Validation Script"
echo "============================="

# Function to check for duplicate column definitions
check_duplicate_columns() {
    echo "Checking for duplicate column definitions..."
    
    # Check if verification fields are defined in multiple migrations
    local verification_migrations=$(grep -r "is_email_verified\|verification_token\|verification_sent_at" migration/db/ --include="*.sql" | wc -l)
    
    if [ "$verification_migrations" -gt 1 ]; then
        echo "⚠️  Warning: Verification fields found in multiple migrations:"
        grep -r "is_email_verified\|verification_token\|verification_sent_at" migration/db/ --include="*.sql" -n
        return 1
    else
        echo "✅ No duplicate column definitions found"
        return 0
    fi
}

# Function to check for missing migration files
check_missing_files() {
    echo "Checking for missing migration files..."
    
    # Get list of migration files
    local migration_files=$(ls migration/db/*.sql | sort)
    
    # Check if any migration files are missing (basic check)
    for file in $migration_files; do
        if [ ! -f "$file" ]; then
            echo "❌ Missing migration file: $file"
            return 1
        fi
    done
    
    echo "✅ All migration files present"
    return 0
}

# Function to validate migration syntax
validate_syntax() {
    echo "Validating migration syntax..."
    
    local has_errors=0
    
    for file in migration/db/*.sql; do
        if [ -f "$file" ]; then
            # Basic SQL syntax check (very basic)
            if ! grep -q "^-- +goose" "$file"; then
                echo "⚠️  Warning: $file may not have proper goose directives"
                has_errors=1
            fi
        fi
    done
    
    if [ $has_errors -eq 0 ]; then
        echo "✅ Migration syntax appears valid"
        return 0
    else
        return 1
    fi
}

# Main validation
main() {
    echo "Starting migration validation..."
    
    local exit_code=0
    
    # Run all checks
    check_duplicate_columns || exit_code=1
    check_missing_files || exit_code=1
    validate_syntax || exit_code=1
    
    if [ $exit_code -eq 0 ]; then
        echo "🎉 All migration validations passed!"
    else
        echo "💥 Migration validation failed!"
        echo "Please fix the issues above before deploying."
    fi
    
    exit $exit_code
}

# Run main function
main "$@" 