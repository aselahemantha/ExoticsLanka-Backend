# Google Cloud Platform Deployment Guide

This guide details how to host the Exotics Lanka backend on Google Cloud Platform (GCP) using Cloud Run, Cloud SQL, and Memorystore (Redis).

## Prerequisites

1.  **GCP Project**: A Google Cloud Project with billing enabled.
2.  **Google Cloud CLI (`gcloud`)**: Installed and authenticated (`gcloud auth login`).
3.  **Docker**: Installed locally for building verify/testing (optional, as Cloud Build handles the actual build).

## 1. Project Setup

Open your terminal and set your project ID:

```bash
export PROJECT_ID="your-gcp-project-id"
gcloud config set project $PROJECT_ID
```

Enable necessary APIs:

```bash
gcloud services enable \
  run.googleapis.com \
  sqladmin.googleapis.com \
  redis.googleapis.com \
  compute.googleapis.com \
  servicenetworking.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com
```

## 2. Infrastructure Setup

### Network Configuration (VPC)
Cloud Run needs to check a Serverless VPC Access connector or use Direct VPC Egress to talk to Cloud SQL and Redis on private IPs. We'll use the default VPC for simplicity.

### Cloud SQL (PostgreSQL)

1.  **Create Instance**:
    ```bash
    gcloud sql instances create exoticslanka-db \
      --database-version=POSTGRES_15 \
      --cpu=2 \
      --memory=4GB \
      --region=us-central1 \
      --root-password=YourStrongPassword123!
    ```

2.  **Create Database**:
    ```bash
    gcloud sql databases create exoticslanka --instance=exoticslanka-db
    ```

3.  **Get Connection Name**:
    ```bash
    export DB_INSTANCE_CONNECTION_NAME=$(gcloud sql instances describe exoticslanka-db --format="value(connectionName)")
    echo "Connection Name: $DB_INSTANCE_CONNECTION_NAME"
    ```

### Memorystore (Redis)

1.  **Create Instance**:
    ```bash
    gcloud redis instances create exoticslanka-redis \
      --size=1 \
      --region=us-central1 \
      --redis-version=redis_7_0
    ```

2.  **Get Host and Port**:
    ```bash
    export REDIS_HOST=$(gcloud redis instances describe exoticslanka-redis --region=us-central1 --format="value(host)")
    export REDIS_PORT=$(gcloud redis instances describe exoticslanka-redis --region=us-central1 --format="value(port)")
    echo "Redis: $REDIS_HOST:$REDIS_PORT"
    ```

## 3. Secret Management

Instead of hardcoding credentials, we use Secret Manager.

1.  **Database URL Secret**:
    Format: `postgres://user:password@/dbname?host=/cloudsql/INSTANCE_CONNECTION_NAME`
    
    *Note: Cloud Run uses a Unix socket to connect to Cloud SQL.*

    ```bash
    # Replace with your actual password and connection name
    echo -n "postgres://postgres:YourStrongPassword123!@/exoticslanka?host=/cloudsql/$DB_INSTANCE_CONNECTION_NAME" | \
    gcloud secrets create exoticslanka-db-url --replication-policy="automatic" --data-file=-
    ```

2.  **Redis URL Secret**:
    ```bash
    echo -n "redis://$REDIS_HOST:$REDIS_PORT" | \
    gcloud secrets create exoticslanka-redis-url --replication-policy="automatic" --data-file=-
    ```

3.  **Grant Access**:
    Ensure the Cloud Run service account has access to these secrets.
    ```bash
    # Get the default Compute Engine service account (used by Cloud Run by default)
    PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')
    SERVICE_ACCOUNT="$PROJECT_NUMBER-compute@developer.gserviceaccount.com"

    # Grant Secret Accessor role
    gcloud secrets add-iam-policy-binding exoticslanka-db-url \
      --member="serviceAccount:$SERVICE_ACCOUNT" \
      --role="roles/secretmanager.secretAccessor"

    gcloud secrets add-iam-policy-binding exoticslanka-redis-url \
      --member="serviceAccount:$SERVICE_ACCOUNT" \
      --role="roles/secretmanager.secretAccessor"
    ```

## 4. Deployment

We have upgraded the `scripts/deploy.sh` script to handle secrets.

### Deploy a Service

To deploy the **auth-service**, run:

```bash
./scripts/deploy.sh auth-service
```

This script will:
1.  Build the container using Cloud Build.
2.  Deploy to Cloud Run in `us-central1`.
3.  Automatically attach the Cloud SQL connection.
4.  Map the secrets `exoticslanka-db-url` and `exoticslanka-redis-url` to environment variables `DATABASE_URL` and `REDIS_URL`.

### Verify Deployment

1.  Go to the [Cloud Run Console](https://console.cloud.google.com/run).
2.  Click on `auth-service`.
3.  Check the URL provided (e.g., `https://auth-service-xyz-uc.a.run.app`).
4.  View logs to ensure it successfully connected to the database and Redis.

## Troubleshooting

-   **Database Connection Failed**:
    -   Ensure `PROJECT_ID` is correct.
    -   Ensure the Cloud SQL Admin API is enabled.
    -   Check that the secret value has the correct connection string format: `host=/cloudsql/<project>:<region>:<instance>`.
    
-   **Redis Connection Failed**:
-   **Redis Connection Failed**:
    -   Ensure `Serverless VPC Access` is configured if using private IP, OR use the default Direct VPC Egress (now standard in newer Cloud Run deployments).

---

## 5. Manual Hosting (Compute Engine / VM)

If you prefer to host everything yourself on a Virtual Machine (e.g., Google Compute Engine, AWS EC2, or DigitalOcean Droplet), follow these steps.

### 1. Provision a VM

Create a VM instance with at least:
-   **OS**: Ubuntu 22.04 LTS (recommended)
-   **CPU**: 2 vCPUs
-   **RAM**: 4GB+ (Go builds and running multiple services can be memory intensive)
-   **Firewall**: Allow HTTP/HTTPS (80/443) and specific service ports (e.g., 8080, 8081) if accessing directly.

**Example GCloud Command:**
```bash
gcloud compute instances create exoticslanka-vm \
    --image-family=ubuntu-2204-lts \
    --image-project=ubuntu-os-cloud \
    --machine-type=e2-medium \
    --tags=http-server,https-server
```

### 2. Install Dependencies

SSH into your VM:
```bash
gcloud compute ssh exoticslanka-vm
```

Update system and install Go, Docker, and Git:

```bash
# Update
sudo apt-get update && sudo apt-get upgrade -y

# Install Git & Make
sudo apt-get install -y git make

# Install Go (adjust version as needed)
wget https://go.dev/dl/go1.25.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.5.linux-amd64.tar.gz
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile
source ~/.profile

# Install Docker
sudo apt-get install -y docker.io docker-compose
sudo usermod -aG docker $USER
# (Log out and back in for group changes to take effect)
```

### 3. Setup Database & Redis

You can install PostgreSQL and Redis directly on the VM or use Docker. We recommend Docker for ease of management.

1.  **Clone Repository**:
    ```bash
    git clone https://github.com/your-username/exoticsLanka.git
    cd exoticsLanka
    ```

2.  **Start Infrastructure**:
    Use the provided `docker-compose.yml` to start Postgres and Redis.
    ```bash
    docker-compose up -d
    ```

3.  **Verify**:
    ```bash
    docker ps
    ```

### 4. Build and Run Services (Systemd)

For a robust manual setup, use `systemd` to keep services running.

1.  **Build Service**:
    ```bash
    cd services/auth-service
    go build -o auth-app ./cmd/api
    ```

2.  **Create Systemd Service File**:
    `sudo nano /etc/systemd/system/auth-service.service`

    ```ini
    [Unit]
    Description=Exotics Lanka Auth Service
    After=network.target docker.service

    [Service]
    User=ubuntu
    ExecStart=/home/ubuntu/exoticsLanka/services/auth-service/auth-app
    WorkingDirectory=/home/ubuntu/exoticsLanka/services/auth-service
    Restart=always
    Environment="DATABASE_URL=postgres://user:password@localhost:5432/exotics_lanka?sslmode=disable"
    Environment="REDIS_URL=redis://localhost:6379"
    Environment="PORT=8081"

    [Install]
    WantedBy=multi-user.target
    ```

3.  **Start Service**:
    ```bash
    sudo systemctl daemon-reload
    sudo systemctl enable auth-service
    sudo systemctl start auth-service
    ```

4.  **Check Status**:
    ```bash
    sudo systemctl status auth-service
    ```

Repeat step 4 for other microservices, ensuring you assign unique ports and update the Nginx/Reverse Proxy configuration if needed.

### 5. Reverse Proxy (Nginx) - Optional

To serve everything on port 80/443:

1.  **Install Nginx**:
    ```bash
    sudo apt-get install -y nginx
    ```

2.  **Configure**:
    Edit `/etc/nginx/sites-available/default` to proxy requests to your Go services based on paths.

    ```nginx
    server {
        listen 80;
        server_name your-domain.com;

        location /api/auth {
            proxy_pass http://localhost:8081;
        }

        location /api/listings {
            proxy_pass http://localhost:8082;
        }
    }
    ```

3.  **Restart Nginx**:
    ```bash
    sudo systemctl restart nginx
    ```
