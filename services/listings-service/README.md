# Listings Service

The Listings Service manages vehicle listings, including creation, retrieval, search, and metrics.

## 🚀 Getting Started

### Prerequisites

-   Go 1.25+
-   PostgreSQL
-   Redis (optional for this service currently, but used by infra)

### Configuration

The service uses environment variables for configuration. Defaults are set for local development.

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8082` | HTTP Server Port |
| `DATABASE_URL` | `postgresql://user:password@localhost:5432/exotics_lanka?sslmode=disable` | Postgres Connection String |

### Running the Service

1.  **Ensure Infrastructure is Up**:
    Make sure Postgres is running (e.g., via Docker Compose in the root directory).
    ```bash
    docker-compose up -d
    ```

2.  **Run the Service**:
    ```bash
    go run cmd/api/main.go
    ```

3.  **Verify**:
    Check if the service is running:
    ```bash
    curl http://localhost:8082/health
    ```

## 🔌 API Endpoints

- `GET /api/listings`: Search listings
- `POST /api/listings`: Create listing
- `GET /api/listings/:id`: Get specific listing
- `GET /api/brands`: Get car brands
