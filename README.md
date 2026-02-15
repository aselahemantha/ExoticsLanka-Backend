# Exotics Lanka Backend

This is the backend repository for the Exotics Lanka platform, built using a modular microservices architecture with Go (Golang).

## 🏗️ Architecture

The project is structured as a Go workspace (monorepo) containing 12 independent microservices.

-   **Language**: Go 1.25+
-   **Database**: PostgreSQL (Primary), Redis (Cache & Session Store)
-   **Infrastructure**: Docker, Google Cloud Run
-   **Communication**: REST API (HTTP)

### Microservices

| Service Name | Description | Port (Local) |
| :--- | :--- | :--- |
| `auth-service` | Authentication, User Management, Sessions | 8081 |
| `listings-service` | Vehicle Listings, Brands, Categories | 8082 |
| `analytics-service` | Platform analytics and metrics | - |
| `comparison-service` | Vehicle comparison functionality | - |
| `contact-service` | Contact forms and inquiries | - |
| `favorites-service` | User favorites and watchlists | - |
| `image-service` | Image processing and storage | - |
| `messaging-service` | User-to-user messaging | - |
| `notification-service`| Push notifications and alerts | - |
| `reports-service` | Data reporting and exports | - |
| `reviews-service` | User reviews and ratings | - |
| `saved-searches` | Saved search preferences | - |

## 🚀 Getting Started

### Prerequisites

-   **Go**: Version 1.25 or higher
-   **Docker** & **Docker Compose**: For running infrastructure (DB, Redis).
-   **Google Cloud SDK**: For deployment.

### Local Development

1.  **Clone the repository**
    ```bash
    git clone <repository-url>
    cd exoticsLanka
    ```

2.  **Start Infrastructure**
    Start PostgreSQL and Redis:
    ```bash
    docker-compose up -d
    ```

3.  **Run a Service**
    Navigate to the service directory or run from root:
    ```bash
    # Example: Run Auth Service
    go run ./services/auth-service/cmd/api/main.go
    ```
    The service will start on its default port (usually defined in `.env` or defaults to 808x).

4.  **Configuration**
    Services use `config.LoadConfig()` to read environment variables.
    -   `DATABASE_URL`: Connection string for PostgreSQL.
    -   `REDIS_URL`: Connection string for Redis.
    -   `PORT`: Port to listen on.

## ☁️ Deployment

The project is configured for deployment on **Google Cloud Run**.

### Prerequisites
-   A Google Cloud Project with billing enabled.
-   `gcloud` CLI authenticated (`gcloud auth login`).

### Deployment Scripts

We provide helper scripts in the `scripts/` directory to automate the build and deploy process.

**1. Deploy a Single Service**
```bash
./scripts/deploy.sh <service-name>
# Example:
./scripts/deploy.sh auth-service
```
This script will:
-   Build the Docker image.
-   Push it to Google Artifact Registry.
-   Deploy the revision to Cloud Run.

**2. Deploy All Services**
```bash
./scripts/deploy_all.sh
```
This will sequentially deploy all services found in the `services/` directory.

### Docker Configuration
Each service has a `Dockerfile` that:
-   Uses a multi-stage build (Go builder -> Alpine runner).
-   Exposes port `8080`.
-   Copies migration files if present.

## 📂 Project Structure

```
.
├── docker-compose.yml       # Local infrastructure (Postgres, Redis)
├── go.work                  # Go workspace configuration
├── scripts/                 # Automation scripts
│   ├── deploy.sh            # Deploy single service
│   ├── deploy_all.sh        # Deploy all services
│   └── generate_dockerfiles.sh # Helper to create Dockerfiles
├── services/                # Microservices source code
│   ├── auth-service/
│   ├── listings-service/
│   └── ... (other services)
└── README.md
```
