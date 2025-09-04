# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Build and Development
- **Build application**: `./scripts/build.sh` (generates SQLC code, runs tests, builds binaries)
- **Run tests**: `go test -v ./...`
- **Run locally**: `go run ./cmd/server`
- **Generate SQLC code**: `sqlc generate`
- **Create migration**: `./scripts/new-migration.sh <migration_name>`

### Build Script Options
The build script supports skipping steps:
- `--skip-sqlc`: Skip SQLC code generation
- `--skip-test`: Skip running tests  
- `--skip-build`: Skip building the application

## Project Architecture

### Core Structure
- **Entry point**: `cmd/server/main.go` - initializes config, database, app, and starts server
- **Application core**: `internal/app.go` - central App struct with database, queries, cache, and email service
- **Configuration**: `internal/config.go` - environment variable handling

### Database Layer
- **Migrations**: `migrations/` directory with Goose migration files
- **SQLC**: Type-safe SQL generation from `internal/db/queries/queries.sql`
- **Generated code**: `internal/db/sqlc/` package (auto-generated, don't edit manually)
- **Database initialization**: `internal/db/db.go` - handles connection and migrations

### API Layer
- **Router**: `internal/api/router.go` - HTTP routes and middleware setup
- **Handlers**: `internal/api/` directory contains endpoint handlers (auth.go, companies.go)
- **Authentication**: JWT-based auth system with password hashing

### Services
- **Email**: `internal/email.go` - email service with configurable API key and from address
- **Caching**: `go-cache` integration with 60-second TTL (hardcoded)

### Testing
- **Test utilities**: `internal/testutil/helpers.go` - shared test helper functions
- **Comprehensive tests**: Each module has corresponding `_test.go` files

## Configuration

### Required Environment Variables
- `DATABASE_URL`: PostgreSQL connection string (required)
- `JWT_SECRET`: JWT signing key (defaults to "dev-secret-change-me")
- `EMAIL_API_KEY`: Email service API key (optional)
- `EMAIL_FROM_ADDRESS`: Sender email address (defaults to "noreply@example.com")

### Database Setup
The application uses PostgreSQL with automatic migrations via Goose. The `internal/db/db.go` package handles database initialization and runs all pending migrations on startup.

## Key Dependencies
- **Go version**: 1.25.0
- **Web framework**: Gin (github.com/gin-gonic/gin)
- **Database**: pgx/v5 for PostgreSQL connectivity
- **Authentication**: golang-jwt/jwt/v5
- **Migrations**: Goose (github.com/pressly/goose/v3)
- **Caching**: go-cache (github.com/patrickmn/go-cache)
- **Type-safe SQL**: SQLC (sqlc.dev)

## Docker Support
The included Dockerfile uses multi-stage builds and includes SQLC code generation as part of the build process. The application runs on port 8080 by default.