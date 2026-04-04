# Journal API

A Go-based monolithic REST API service with authentication, user management, and JWT token handling. Built with Clean Architecture principles for maintainability and scalability.

## 🚀 Tech Stack

- **Language:** Go 1.24.2
- **Framework:** Gin (HTTP)
- **ORM:** GORM
- **Database:** PostgreSQL
- **Cache:** Redis
- **Authentication:** JWT (Access + Refresh Token)
- **Migration:** Goose
- **Config:** Viper (Environment Variables)

## 📋 Prerequisites

- Go 1.24.2 or higher
- PostgreSQL 12+
- Redis 6+
- Make (optional, for using Makefile commands)

## ⚙️ Configuration

This application uses **environment variables** for all configuration. No YAML/config files needed!

### Development Setup

1. **Copy the environment template:**
   ```bash
   cp env.example .env
   ```

2. **Edit `.env` with your local settings:**
   ```bash
   # Required settings for local development
   ENV=development
   LOG_LEVEL=debug
   
   # Database (update with your PostgreSQL credentials)
   DATABASE_DSN=postgres://postgres:postgres@localhost:5432/journal_api?sslmode=disable
   
   # Redis (update with your Redis settings)
   REDIS_CACHE_DSN=redis://localhost:6379/0
   REDIS_IS_CACHE_DISABLE=false
   
   # JWT Secrets (generate secure random strings for production!)
   TOKEN_PASSWORD_SALT=your-16plus-chars-salt
   TOKEN_ACCESS_TOKEN_SECRET=your-16plus-chars-secret
   TOKEN_REFRESH_TOKEN_SECRET=your-16plus-chars-secret
   ```

3. **Generate secure secrets for production:**
   ```bash
   # On Linux/Mac:
   openssl rand -base64 32
   
   # Or:
   head -c 32 /dev/urandom | base64
   ```

### Production Deployment

Set environment variables in your deployment platform (Railway, Render, Kubernetes, etc.):

```bash
ENV=production
LOG_LEVEL=info
DATABASE_DSN=postgres://user:password@prod-host:5432/prod_db?sslmode=require
REDIS_CACHE_DSN=redis://:password@redis-host:6379/0
TOKEN_PASSWORD_SALT=$(openssl rand -base64 32)
TOKEN_ACCESS_TOKEN_SECRET=$(openssl rand -base64 32)
TOKEN_REFRESH_TOKEN_SECRET=$(openssl rand -base64 32)
```

### Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ENV` | No | `development` | Environment: development, staging, production |
| `LOG_LEVEL` | No | `info` | Log level: debug, info, warn, error |
| `SERVER_PORT` | No | `8080` | HTTP server port |
| `DATABASE_DSN` | **Yes** | - | PostgreSQL connection string |
| `REDIS_CACHE_DSN` | **Yes*** | - | Redis connection string (*required if cache enabled) |
| `REDIS_IS_CACHE_DISABLE` | No | `false` | Set to `true` to disable caching |
| `TOKEN_PASSWORD_SALT` | **Yes** | - | Password hashing salt (min 16 chars) |
| `TOKEN_ACCESS_TOKEN_SECRET` | **Yes** | - | JWT access token secret (min 16 chars) |
| `TOKEN_REFRESH_TOKEN_SECRET` | **Yes** | - | JWT refresh token secret (min 16 chars) |
| `TOKEN_ACCESS_TOKEN_DURATION` | No | `1h` | Access token expiration (e.g., 15m, 1h) |
| `TOKEN_REFRESH_TOKEN_DURATION` | No | `720h` | Refresh token expiration (e.g., 24h, 720h) |

See `env.example` for complete list with advanced settings.

## 🛠️ Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd journal-api
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Setup environment variables:**
   ```bash
   cp env.example .env
   # Edit .env with your configuration
   ```

4. **Start PostgreSQL and Redis:**
   ```bash
   # Using Docker Compose (if available)
   docker-compose up -d postgres redis
   
   # Or start them manually
   ```

5. **Run database migrations:**
   ```bash
   go run main.go migrate up
   # Or using make:
   make migrate-up
   ```

6. **Start the server:**
   ```bash
   go run main.go server
   # Or using make:
   make run
   ```

The API will be available at `http://localhost:8080`

## 📦 Available Commands

```bash
# Run the server
go run main.go server

# Run migrations
go run main.go migrate up
go run main.go migrate down

# Using Makefile (if available)
make run              # Run the server
make migrate-up       # Run all migrations
make migrate-down     # Rollback last migration
make migrate-status   # Check migration status
make build           # Build binary
make test            # Run tests
```

## 📖 API Documentation

### Response Format

All API responses follow this structure:

```json
{
	"message": "success message or error message",
	"data": {
		// Response data (object, array, or null)
	},
	"validationErrors": [
		{
			"field": "fieldname",
			"tag": "validation_rule",
			"message": "error message"
		}
	]
}
```

### Authentication Endpoints

#### 1. Register

Create a new user account.

**Endpoint:** `POST /v1/auth/register`

**Request Body:**
```json
{
	"username": "username",
	"email": "email@gmail.com",
	"password": "strongpassword"
}
```

**Success Response:**
```json
{
	"message": "ok",
	"data": null,
	"validationErrors": null
}
```

**Error Response:**
```json
{
	"message": "validation error",
	"data": null,
	"validationErrors": [
		{
			"field": "username",
			"tag": "unique_db",
			"message": "username already taken"
		}
	]
}
```

#### 2. Login

Authenticate user and receive JWT tokens.

**Endpoint:** `POST /v1/auth/login`

**Request Body:**
```json
{
	"identifier": "username or email",
	"password": "strongpassword"
}
```

**Success Response:**
```json
{
	"message": "ok",
	"data": {
		"accessToken": "eyJhbGci...",
		"accessTokenExpiredAt": "2024-06-17T09:23:04Z",
		"refreshToken": "eyJhbGci...",
		"refreshTokenExpiredAt": "2024-06-17T09:23:04Z"
	},
	"validationErrors": null
}
```

#### 3. Refresh Token

Get new access token using refresh token.

**Endpoint:** `POST /v1/auth/refresh`

**Headers:**
```
Authorization: Bearer <REFRESH_TOKEN>
```

**Success Response:**
```json
{
	"message": "ok",
	"data": {
		"accessToken": "eyJhbGci...",
		"accessTokenExpiredAt": "2024-06-17T09:23:04Z",
		"refreshToken": "eyJhbGci...",
		"refreshTokenExpiredAt": "2024-06-17T09:23:04Z"
	},
	"validationErrors": null
}
```

#### 4. User Info

Get current authenticated user information.

**Endpoint:** `GET /v1/auth/info`

**Headers:**
```
Authorization: Bearer <ACCESS_TOKEN>
```

**Success Response:**
```json
{
	"message": "ok",
	"data": {
		"id": "6759e56a-fa7c-49ee-9854-f32ab38083ae",
		"username": "username",
		"firstName": "First",
		"lastName": "Last",
		"email": "email@gmail.com",
		"phone": "1234567890",
		"location": "Jakarta",
		"createdAt": "2024-06-16T10:46:25Z",
		"updatedAt": "2024-06-16T10:46:25Z"
	},
	"validationErrors": null
}
```

#### 5. Logout

Invalidate current access token.

**Endpoint:** `POST /v1/auth/logout`

**Headers:**
```
Authorization: Bearer <ACCESS_TOKEN>
```

**Success Response:**
```json
{
	"message": "ok",
	"data": null,
	"validationErrors": null
}
```

## 🏗️ Project Structure

```
journal-api/
├── cmd/                          # CLI commands
│   ├── root.go                  # Root command
│   ├── server.go                # Server command
│   └── migrate.go               # Migration command
├── internal/                     # Private application code
│   ├── bootstrap/               # Dependency injection
│   ├── config/                  # Configuration management
│   ├── constant/                # Constants and error definitions
│   ├── infrastructure/          # External dependencies (DB, Redis, Gin)
│   ├── model/
│   │   ├── entity/             # Domain entities
│   │   ├── request/            # API request DTOs
│   │   ├── response/           # API response DTOs
│   │   ├── contract/           # Interface definitions
│   │   └── cachekey/           # Cache key patterns
│   ├── repository/              # Data access layer
│   │   ├── cache/              # Cache repository
│   │   └── user/               # User repository
│   ├── service/                 # Business logic layer
│   │   └── auth/               # Authentication service
│   └── transport/               # Presentation layer
│       └── http/               # HTTP handlers & routes
├── migration/db/                # SQL migration files
├── .env                         # Environment variables (git-ignored)
├── env.example                  # Environment variables template
├── Dockerfile                   # Docker build configuration
├── docker-compose.yml          # Docker Compose setup
└── main.go                     # Application entry point
```

## 🔧 Development

### Adding New Features

Follow Clean Architecture principles:

1. **Define contracts** in `internal/model/contract/`
2. **Create entity models** in `internal/model/entity/`
3. **Implement repository** in `internal/repository/<domain>/`
4. **Create service layer** in `internal/service/<domain>/`
5. **Add HTTP handlers** in `internal/transport/http/<domain>/`
6. **Register routes** in `internal/transport/http/route.transport.go`
7. **Wire dependencies** in `internal/bootstrap/`

See the existing `auth` implementation as a reference.

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Database Migrations

Create a new migration:

```bash
# Using goose directly
goose -dir migration/db create migration_name sql

# Or using the app
go run main.go migrate create migration_name
```

**DB lama / skema sudah ada tanpa riwayat Goose:** migrasi di `migration/db` diset **idempoten** (`IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`, dll.) supaya `migrate up` tidak gagal bila tabel/kolom sudah dibuat lewat seed atau eksperimen. Alternatif lain tanpa mengubah SQL: baseline Goose (tandai versi sudah terpasang), mis. `go run main.go migrate status` lalu isi tabel `goose_db_version` sesuai dokumen [pressly/goose](https://github.com/pressly/goose), atau untuk dev paling bersih: **database kosong baru** lalu `migrate up` sekali.

## 🐳 Docker

### Build and Run with Docker

```bash
# Build image
docker build -t journal-api .

# Run container
docker run -p 8080:8080 \
  -e DATABASE_DSN="postgres://..." \
  -e REDIS_CACHE_DSN="redis://..." \
  -e TOKEN_PASSWORD_SALT="..." \
  -e TOKEN_ACCESS_TOKEN_SECRET="..." \
  -e TOKEN_REFRESH_TOKEN_SECRET="..." \
  journal-api
```

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Stop all services
docker-compose down
```

## 🚀 Deployment

### Railway

1. Create new project on Railway
2. Add PostgreSQL and Redis services
3. Set environment variables from `env.example`
4. Deploy from GitHub repository

### Render

1. Create new Web Service
2. Add PostgreSQL and Redis add-ons
3. Set environment variables
4. Deploy

### Kubernetes

See `deployments/` directory for Kubernetes manifests (if available).

## 📝 License

[Your License Here]

## 🤝 Contributing

Contributions are welcome! Please follow the existing code structure and conventions.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📞 Support

For issues and questions, please open an issue on GitHub.
