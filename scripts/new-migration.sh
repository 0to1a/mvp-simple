#!/bin/bash

# Script to create new database migration files
# Usage: ./scripts/new-migration.sh migration_name

set -e

# Check if migration name is provided
if [ -z "$1" ]; then
    echo "Error: Migration name is required"
    echo "Usage: $0 <migration_name>"
    echo "Example: $0 add_user_table"
    exit 1
fi

# Create migrations directory if it doesn't exist
mkdir -p migrations

# Generate timestamp and migration name
TIMESTAMP=$(date +%s)
MIGRATION_NAME="$1"
MIGRATION_FILE="migrations/${TIMESTAMP}_${MIGRATION_NAME}.sql"

# Create migration file with template
cat > "$MIGRATION_FILE" << EOF
-- +goose Up
-- SQL for forward migration goes here


-- +goose Down
-- SQL for rollback migration goes here

EOF

echo "✅ Created migration: $MIGRATION_FILE"
echo "📝 Edit the file to add your SQL statements"
