# MVP Backend Simple

A minimal Go backend API with authentication and PostgreSQL integration.

## Features

- JWT authentication with secure password hashing
- PostgreSQL with SQLC for type-safe queries
- Database migrations with Goose
- RESTful API with Gin framework
- Docker support

## Tech Stack

- Go 1.25
- Gin (web framework)
- PostgreSQL 16
- SQLC (type-safe SQL)
- Goose (migrations)
- JWT (golang-jwt/jwt/v5)

## Quick Start

1. **Clone and setup**
   ```bash
   git clone <repository-url>
   cd mvp-backend-simple
   cp .env.example .env
   ```

2. **Start database**
   ```bash
   # Using Docker
   docker run -d --name mvp-postgres \
     -e POSTGRES_USER=mvp \
     -e POSTGRES_PASSWORD=mvp \
     -e POSTGRES_DB=mvp \
     -p 5432:5432 \
     postgres:16
   ```

3. **Run the app**
   ```bash
   go run ./cmd/server
   ```

Server starts on `http://localhost:8080`

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes |
| `JWT_SECRET` | JWT secret key | No |

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
go test ./...
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
