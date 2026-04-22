# Exotics Lanka Modular Monolith

This repository contains the backend for the Exotics Lanka platform, built as a **Modular Monolith** using Go (Golang). It combines the benefits of modularity with the simplicity of a single deployment unit.

## 🏗️ Architecture

The project is structured as a modular monolith, where each business domain (Auth, Listings, Messaging, etc.) is isolated within the `internal/` directory.

-   **Language**: Go 1.25+
-   **Framework**: Gin Gonic (HTTP)
-   **Database**: PostgreSQL (Primary), Redis (Cache & Session Store)
-   **Infrastructure**: Docker, Google Cloud Run
-   **Design Pattern**: Clean Architecture (Domain-Driven Design)

### Modules

The following modules are integrated into the monolith:

| Module | Description |
| :--- | :--- |
| **Auth** | Authentication, RBAC, Sessions, Password Recovery |
| **User** | Profile management, Seller verification |
| **Listings** | Vehicle inventory, Search, Filtering |
| **Image** | Cloudinary integration, Image reordering |
| **Messaging** | Real-time chat between buyers and sellers |
| **Notification**| Email (SendGrid) and SMS (Twilio) alerts |
| **Reviews** | Ratings and feedback for sellers |
| **Favorites** | User watchlists |
| **Comparison** | Vehicle side-by-side comparison |
| **Contact** | Inquiry management and lead tracking |
| **Reports** | Moderation and listing reporting |
| **Saved Searches**| Automated matching for user preferences |
| **Analytics** | Platform-wide metrics and daily aggregations |

## 🚀 Getting Started

### Prerequisites

-   **Go**: Version 1.25 or higher
-   **Docker** & **Docker Compose**: For local infrastructure.
-   **GCP SDK**: For cloud deployment.

### Local Development

1.  **Start Infrastructure**
    Launch PostgreSQL and Redis using Docker Compose:
    ```bash
    docker-compose up -d
    ```

2.  **Run the Monolith**
    ```bash
    go run ./cmd/api/main.go
    ```
    The server will start on port `8080` by default.

3.  **Testing with HTTP Client**
    Use [request.http](./request.http) in your IDE to test endpoints. The file includes scripts to automatically capture and reuse the `access_token` after login.

4.  **Testing with Postman**
    Import [ExoticsLanka_Monolith.postman_collection.json](./ExoticsLanka_Monolith.postman_collection.json) for a comprehensive suite of API tests.

## ☁️ Deployment

The project is deployed to **Google Cloud Run** using Google Cloud Build.

### Deployment Command
```bash
gcloud builds submit --config cloudbuild.yaml .
```

### Docker
The monolith uses a multi-stage Docker build for optimized image size.
-   **Build Stage**: Compiles the Go binary.
-   **Run Stage**: Alpine-based lightweight image.

## 📂 Project Structure

```
.
├── cmd/api/main.go          # Entry point for the Monolith
├── internal/                # Modular domains
│   ├── auth/                # Auth logic
│   ├── listings/            # Listings logic
│   └── ...                  # Other modules
├── migrations/              # SQL migration files per module
├── request.http             # Comprehensive API test file
├── cloudbuild.yaml          # Google Cloud Build config
├── docker-compose.yml       # Local dev infrastructure
└── README.md
```
