# MVP Backend Simple

A minimal, production-ready Go backend API with authentication, database integration, and clean architecture.

## Features

- **Authentication**: JWT-based authentication with secure password hashing
- **Database**: PostgreSQL integration with type-safe queries using SQLC
- **Migrations**: Database schema management with Goose
- **Caching**: In-memory caching with configurable TTL
- **API**: RESTful API built with Gin framework
- **Docker**: Containerized development environment
- **Clean Architecture**: Well-organized project structure following Go conventions

## Tech Stack

- **Language**: Go 1.23
- **Web Framework**: Gin
- **Database**: PostgreSQL 16
- **Query Builder**: SQLC (type-safe SQL)
- **Migrations**: Goose
- **Authentication**: JWT with golang-jwt/jwt/v5
- **Caching**: patrickmn/go-cache
- **Containerization**: Docker

## Project Structure

```
mvp-backend-simple/
├── cmd/server/           # Main application entry point
│   └── main.go
├── internal/            # Private application code
│   ├── api/            # API handlers and routing
│   ├── auth/           # Authentication logic
│   ├── db/             # Database layer
│   │   ├── queries/    # SQL queries for SQLC
│   │   └── sqlc/       # Generated SQLC code
│   ├── app.go          # Application setup and dependency injection
│   └── config.go       # Configuration management
├── migrations/          # Database migration files
│   └── 0001_init.sql
├── scripts/            # Development scripts
│   ├── new-migration.sh # Create new database migrations
│   └── build.sh        # Generate SQLC, test, and build
├── Dockerfile          # Container build configuration
├── .dockerignore       # Docker build context exclusions
├── go.mod              # Go module dependencies
├── sqlc.yaml           # SQLC configuration
└── .env.example        # Environment variables template
```

## Quick Start

### Prerequisites

- Go 1.23 or later
- Docker (optional, for containerized deployment)
- PostgreSQL (local installation or remote instance)
- [SQLC](https://docs.sqlc.dev/en/stable/overview/install.html) (for code generation)

### 1. Clone and Setup

```bash
git clone <repository-url>
cd mvp-backend-simple
```

### 2. Environment Configuration

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 3. Setup Database

**Option A: Local PostgreSQL**
```bash
# Install PostgreSQL locally and create database
creatdb mvp
```

**Option B: Docker PostgreSQL**
```bash
# Run PostgreSQL in Docker
docker run -d \
  --name mvp-postgres \
  -e POSTGRES_USER=mvp \
  -e POSTGRES_PASSWORD=mvp \
  -e POSTGRES_DB=mvp \
  -p 5432:5432 \
  postgres:16
```

### 4. Generate Code (if needed)

```bash
sqlc generate
```

### 5. Run the Application

```bash
go run ./cmd/server
```

The server will start on `http://localhost:8080`

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | - | Yes |
| `JWT_SECRET` | Secret key for JWT tokens | `dev-secret-change-me` | No |

## Development

### Database Operations

#### Creating New Migrations

**Using the script (recommended):**
```bash
# Create a new migration with template
./scripts/new-migration.sh add_user_table
./scripts/new-migration.sh create_posts_table
```

**Manual method:**
```bash
# Create a new migration file manually
TIMESTAMP=$(date +%s)
MIGRATION_NAME="your_migration_name"
touch migrations/${TIMESTAMP}_${MIGRATION_NAME}.sql
```

#### Full Development Build

**Using the build script (recommended):**
```bash
# Full build: SQLC generation + tests + build
./scripts/build.sh

# Skip specific steps
./scripts/build.sh --skip-test     # Skip tests
./scripts/build.sh --skip-sqlc     # Skip SQLC generation
./scripts/build.sh --skip-build    # Skip build step

# Show help
./scripts/build.sh --help
```

**Manual SQLC generation:**
```bash
# Generate SQLC code only
sqlc generate
```

### Building

**Local Build:**
```bash
# Build the application
go build -o bin/server ./cmd/server

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/server-linux ./cmd/server
```

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Vet code
go vet ./...
```

## Database Schema

The application uses PostgreSQL with the following main tables:
- `users` - User accounts and authentication
- Additional tables as defined in migrations

## Configuration Details

### Database Connection
The application expects a PostgreSQL connection string in the format:
```
postgres://username:password@host:port/database?sslmode=disable
```

### JWT Configuration
- Tokens are signed using the `JWT_SECRET` environment variable
- Default expiration and refresh logic is implemented
- Secure password hashing using bcrypt

### Caching
- In-memory cache with configurable TTL
- Used for session management and performance optimization

## Production Deployment

### Environment Setup
1. Set secure `JWT_SECRET` (use a strong, random string)
2. Configure production PostgreSQL instance

**Note**: The application runs on port 8080 (hardcoded) and uses a 60-second cache TTL (hardcoded).

### Security Considerations
- Change default JWT secret in production
- Use SSL/TLS for database connections in production
- Implement rate limiting (consider adding middleware)
- Use environment-specific configurations

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run `go vet ./...` and ensure tests pass
6. Submit a pull request

## License

[Add your license here]

## Support

For questions or issues, please [create an issue](https://github.com/your-repo/issues) in the repository.
