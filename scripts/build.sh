#!/bin/bash

# Script to generate SQLC code, run tests, and build the application
# Usage: ./scripts/build.sh [options]
# Options:
#   --skip-sqlc    Skip SQLC generation
#   --skip-test    Skip running tests
#   --skip-build   Skip building the application

set -e

SKIP_SQLC=false
SKIP_TEST=false
SKIP_BUILD=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-sqlc)
            SKIP_SQLC=true
            shift
            ;;
        --skip-test)
            SKIP_TEST=true
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --skip-sqlc    Skip SQLC generation"
            echo "  --skip-test    Skip running tests"
            echo "  --skip-build   Skip building the application"
            echo "  -h, --help     Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

echo "🚀 Starting build process..."

# Step 1: Generate SQLC code
if [ "$SKIP_SQLC" = false ]; then
    echo "📝 Generating SQLC code..."
    if ! command -v sqlc &> /dev/null; then
        echo "❌ Error: SQLC is not installed"
        echo "Install SQLC: https://docs.sqlc.dev/en/stable/overview/install.html"
        exit 1
    fi
    sqlc generate
    echo "✅ SQLC code generated successfully"
else
    echo "⏭️  Skipping SQLC generation"
fi

# Step 2: Run tests
if [ "$SKIP_TEST" = false ]; then
    echo "🧪 Running tests..."
    go mod tidy
    go vet ./...
    go test -v ./...
    echo "✅ Tests passed successfully"
else
    echo "⏭️  Skipping tests"
fi

# Step 3: Build application
if [ "$SKIP_BUILD" = false ]; then
    echo "🔨 Building application..."
    
    # Create bin directory if it doesn't exist
    mkdir -p bin
    
    # Build for current platform
    go build -o bin/server ./cmd/server
    echo "✅ Application built successfully: bin/server"
    
    # Build for Linux (useful for Docker/deployment)
    echo "🐧 Building for Linux..."
    GOOS=linux GOARCH=amd64 go build -o bin/server-linux ./cmd/server
    echo "✅ Linux binary built successfully: bin/server-linux"
else
    echo "⏭️  Skipping build"
fi

echo "🎉 Build process completed successfully!"

# Display binary info
if [ "$SKIP_BUILD" = false ]; then
    echo ""
    echo "📊 Build artifacts:"
    ls -lah bin/
fi
