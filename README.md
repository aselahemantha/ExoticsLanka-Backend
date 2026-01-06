# Exotics Lanka Backend

This is the backend repository for the Exotics Lanka platform, built using a microservices architecture with Go.

## 🏗️ Architecture

The project follows a modular microservices architecture managed as a Go workspace (monorepo).

- **API Gateway**: (Planned)
- **Services**:
  - `auth-service`: Authentication and user management.
- **Infrastructure**:
  - PostgreSQL (Primary Database)
  - Redis (Caching & Sessions)

For detailed architectural decisions, please refer to:
- [Microservices Architecture](MICROSERVICES_ARCHITECTURE.md)
- [API Specifications](API_SPECIFICATIONS.md)
- [Implementation Roadmap](BACKEND_IMPLEMENTATION_ROADMAP.md)

## 🚀 Getting Started

### Prerequisites

- **Go**: Version 1.25 or higher
- **Docker**: For running database and cache services

### Installation

1.  **Clone the repository**
    ```bash
    git clone <repository-url>
    cd exoticsLanka
    ```

2.  **Initialize Infrastructure**
    Start PostgreSQL and Redis using Docker Compose:
    ```bash
    docker-compose up -d
    ```

3.  **Run Services**
    
    **Auth Service**:
    ```bash
    go run ./services/auth-service/cmd/api/main.go
    ```
    The service will start on port `8081`. You can check the health at `http://localhost:8081/health`.

## 📂 Project Structure

```
.
├── docker-compose.yml   # Infrastructure definition (Postgres, Redis)
├── go.work              # Go workspace configuration
├── services/            # Microservices directory
│   └── auth-service/    # Authentication Service
│       ├── cmd/         # Entry points
│       └── internal/    # Private application code
├── API_SPECIFICATIONS.md
├── MICROSERVICES_ARCHITECTURE.md
└── BACKEND_IMPLEMENTATION_ROADMAP.md
```

## 📅 Current Status

**Phase 1: Foundation**
- [x] Project Structure (Go Workspace)
- [x] Docker Infrastructure (Postgres, Redis)
- [x] Auth Service Scaffolding
- [ ] Auth Service Implementation (In Progress)
