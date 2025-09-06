# MVP Simple API

A complete Go backend API with JWT authentication, user management, company management, and role-based access control.

## Features

- **JWT Authentication**: Secure OTP-based authentication with refresh tokens
- **User Management**: Admin-only user creation, listing, and soft deletion
- **Company Management**: Multi-company support with user assignments
- **Role-Based Access Control**: Admin vs regular user permissions
- **Development Mode**: Mock authentication for testing (test@test.com / 123456)
- **PostgreSQL Integration**: Type-safe queries with SQLC
- **Database Migrations**: Automatic schema management with Goose
- **RESTful API**: Clean endpoints with Gin framework
- **Docker Support**: Multi-stage builds and containerization

## Tech Stack

- Go 1.25
- Gin (web framework)
- PostgreSQL 16
- SQLC (type-safe SQL)
- Goose (migrations)
- JWT (golang-jwt/jwt/v5)
- pgx/v5 (PostgreSQL driver)

## Quick Start

1. **Clone and setup**
   ```bash
   git clone <repository-url>
   cd mvp-backend-simple
   cp .env.example .env
   ```

2. **Configure environment**
   ```bash
   # Required
   export DATABASE_URL="postgres://user:password@localhost:5432/dbname"
   
   # Optional - defaults provided
   export JWT_SECRET="your-jwt-secret"
   export ENVIRONMENT="dev"  # or "prod"
   export EMAIL_API_KEY="your-email-api-key"
   export EMAIL_FROM_ADDRESS="noreply@example.com"
   ```

3. **Start database**
   ```bash
   # Using Docker
   docker run -d --name mvp-postgres \
     -e POSTGRES_USER=mvp \
     -e POSTGRES_PASSWORD=mvp \
     -e POSTGRES_DB=mvp \
     -p 5432:5432 \
     postgres:16
   ```

4. **Run the app**
   ```bash
   go run ./cmd/server
   ```

Server starts on `http://localhost:8080`

## Development Mode

When `ENVIRONMENT=dev` (default), the API provides special development features:

- **Test Account**: `test@test.com` with OTP `123456`
- **Mock OTP**: Always uses `123456` for test@test.com
- **Auto-seeding**: Test user and company are automatically created
- **No Email**: Skips actual email sending for test@test.com

**Example usage:**
```bash
# Request OTP (returns mock OTP in dev mode)
curl -X POST http://localhost:8080/v1/login/request \
  -H "Content-Type: application/json" \
  -d '{"email": "test@test.com"}'

# Login with mock OTP
curl -X POST http://localhost:8080/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@test.com", "otp": "123456"}'
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `DATABASE_URL` | PostgreSQL connection string | - | Yes |
| `JWT_SECRET` | JWT signing key | `dev-secret-change-me` | No |
| `ENVIRONMENT` | Runtime environment | `dev` | No |
| `EMAIL_API_KEY` | Email service API key | - | No |
| `EMAIL_FROM_ADDRESS` | Sender email address | `noreply@example.com` | No |

## Development

### Create migration
```bash
./scripts/new-migration.sh migration_name
```

### Build project
```bash
./scripts/build.sh
```

### Run tests
```bash
go test -v ./...
```

## Project Structure

```
├── cmd/server/      # Application entry point
├── internal/        # Private code
│   ├── api/         # Handlers
│   ├── auth/        # Authentication
│   └── db/          # Database layer
├── migrations/      # DB migrations
└── scripts/         # Development scripts
```
