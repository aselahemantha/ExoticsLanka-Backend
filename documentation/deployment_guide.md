# Google Cloud Platform Deployment Guide

This guide details how to host the Exotics Lanka backend (Auth, Listings, and Image services) on Google Cloud Platform (GCP) using Cloud Run, Cloud SQL, and Memorystore (Redis).

## Prerequisites

1.  **GCP Project**: A Google Cloud Project with billing enabled.
2.  **Google Cloud CLI (`gcloud`)**: Installed and authenticated (`gcloud auth login`).

## 1. Project & API Setup

Set your project ID and enable the mandatory APIs:

```bash
export PROJECT_ID="your-project-id"
gcloud config set project $PROJECT_ID

gcloud services enable \
  run.googleapis.com \
  sqladmin.googleapis.com \
  redis.googleapis.com \
  compute.googleapis.com \
  vpcaccess.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com
```

## 2. Infrastructure Setup (Step-by-Step)

You must create the database, cache, and networking before deploying the services.

### A. Cloud SQL (PostgreSQL)
1.  **Create Instance**:
    ```bash
    gcloud sql instances create exotics-lanka-db \
      --database-version=POSTGRES_15 \
      --tier=db-f1-micro \
      --region=us-central1 \
      --root-password="your-db-password"
    ```
2.  **Create Database**:
    ```bash
    gcloud sql databases create exotics_lanka --instance=exotics-lanka-db
    ```
3.  **Get Connection Name**:
    ```bash
    gcloud sql instances describe exotics-lanka-db --format="value(connectionName)"
    # Example output: exotics-lanka:us-central1:exotics-lanka-db
    ```

### B. Cloud Memorystore (Redis)
1.  **Create Instance**:
    ```bash
    gcloud redis instances create exotics-lanka-redis \
      --size=1 --region=us-central1 --redis-version=redis_7_0
    ```
2.  **Get Host IP**:
    ```bash
    gcloud redis instances describe exotics-lanka-redis --region=us-central1 --format='value(host)'
    ```

### C. VPC Access Connector (Mandatory)
Required for Cloud Run to talk to the private SQL and Redis instances.
```bash
gcloud compute networks vpc-access connectors create exotics-lanka-connector \
  --region=us-central1 \
  --range=10.8.0.0/28
```

### D. Artifact Registry
Create a repository to store your Docker images:
```bash
gcloud artifacts repositories create exotics-lanka-repo \
  --repository-format=docker \
  --location=us-central1
```

## 3. IAM Permissions

Both Cloud Build (for deploying) and Cloud Run (for running) need access to Secret Manager.

```bash
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')

# Grant to Cloud Build service account
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$PROJECT_NUMBER@cloudbuild.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"

# Grant to Cloud Run runtime service account
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

## 4. Secret Manager Configuration

Create the secrets and add your actual connection strings.

1.  **Create Secret Containers**:
    ```bash
    gcloud secrets create DATABASE_URL --replication-policy="automatic"
    gcloud secrets create REDIS_URL --replication-policy="automatic"
    gcloud secrets create JWT_SECRET --replication-policy="automatic"
    gcloud secrets create CLOUDINARY_URL --replication-policy="automatic"
    ```

2.  **Add Values**:
    -   **DATABASE_URL**: `postgres://postgres:your-password@/exotics_lanka?host=/cloudsql/PROJECT_ID:us-central1:exotics-lanka-db`
    -   **REDIS_URL**: `redis://<REDIS_IP>:6379`
    
    ```bash
    echo -n "your-connection-string" | gcloud secrets versions add DATABASE_URL --data-file=-
    # Repeat for others...
    ```

## 5. CI/CD with Cloud Build Triggers

We use the parameterized `cloudbuild.yaml` at the root of the project.

### Setup Triggers
Create a trigger in the [GCP Console](https://console.cloud.google.com/cloud-build/triggers) for each service:

| Service | Included Files Filter | _SERVICE | _SERVICE_PATH |
| :--- | :--- | :--- | :--- |
| Auth | `services/auth-service/**` | `auth-service` | `services/auth-service` |
| Listings | `services/listings-service/**` | `listings-service` | `services/listings-service` |
| Image | `services/image-service/**` | `image-service` | `services/image-service` |

**Trigger Configuration Details:**
-   **Event**: Push to branch (e.g., `main`).
-   **Configuration Type**: Cloud Build configuration file (yaml or json).
-   **Cloud Build configuration file location**: `cloudbuild.yaml`.
-   **Advanced / Substitutions**: Add the variables listed in the table above.

## 6. Verification

After a trigger runs:
1.  Check the **Cloud Build** history for a successful build.
2.  Visit the **Cloud Run** service URL (found in the console).
3.  Check logs for "Connected to PostgreSQL" and "Connected to Redis" messages.
