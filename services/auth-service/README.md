# Auth Service

The Auth Service is a critical microservice in the Exotics Lanka platform, responsible for handling user authentication, authorization, and session management. It is built using Go and follows the Clean Architecture pattern to ensure scalability, maintainability, and testability.

## 🏗 Architecture

The project is structured into layers:

-   **`cmd/api`**: Entry point (`main.go`). Initializes dependencies and starts the server.
-   **`internal/config`**: Configuration loading using environment variables.
-   **`internal/domain`**: Pure domain models and interfaces. This layer has no external dependencies.
-   **`internal/repository`**: Data persistence implementations for PostgreSQL and Redis.
-   **`internal/usecase`**: Core business logic (Registration, Login, JWT generation, Password hashing).
-   **`internal/delivery/http`**: HTTP Handlers and Middleware using the Gin framework.

## 🛠 Tech Stack & Packages

-   **Web Framework**: `github.com/gin-gonic/gin`
-   **Database Driver**: `github.com/jackc/pgx/v5` (PostgreSQL)
-   **Cache Client**: `github.com/redis/go-redis/v9` (Redis)
-   **Authentication**: `github.com/golang-jwt/jwt/v5` (JWT)
-   **Password Hashing**: `golang.org/x/crypto/bcrypt`
-   **Context & Utils**: `github.com/google/uuid`, `github.com/joho/godotenv`

## 🚀 Getting Started

### Prerequisites

-   Go 1.22+
-   PostgreSQL
-   Redis

### Configuration

Create a `.env` file in the root of the service:
```env
PORT=8081
DATABASE_URL=postgres://user:password@localhost:5432/exotics_lanka?sslmode=disable
REDIS_URL=redis://localhost:6379
JWT_SECRET=your_jwt_secret_key
JWT_REFRESH_SECRET=your_refresh_secret_key
```

### Running the Service

1.  **Run Migrations**: Execute the SQL scripts in `sql/migrations/` to setup your database schema.
2.  **Start the Server**:
    ```bash
    go run cmd/api/main.go
    ```

## 🔌 API Endpoints

### Public Routes
-   `POST /api/auth/register`: Register a new user
-   `POST /api/auth/login`: Login with email and password
-   `POST /api/auth/refresh`: Refresh access token
-   `POST /api/auth/verify-email`: Verify email address
-   `POST /api/auth/forgot-password`: Request password reset
-   `POST /api/auth/reset-password`: Reset password with token

### Protected Routes (Requires Header `Authorization: Bearer <token>`)
-   `GET /api/auth/me`: Get current user profile
-   `POST /api/auth/logout`: Logout user
-   `POST /api/auth/change-password`: Change password

## 🧪 Testing

You can use the provided `requests.http` file to test endpoints directly in your IDE (JetBrains/VSCode).

Or use curl:
```bash
curl -X GET http://localhost:8081/api/auth/me -H "Authorization: Bearer <token>"
```
