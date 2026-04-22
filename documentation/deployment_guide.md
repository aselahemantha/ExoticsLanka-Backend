# Deployment Guide (Modular Monolith)

This guide details how to deploy the Exotics Lanka modular monolith to Google Cloud Platform (GCP).

## 1. Project Setup

Enable mandatory APIs:
```bash
gcloud services enable \
  run.googleapis.com \
  sqladmin.googleapis.com \
  redis.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com
```

## 2. Infrastructure

Ensure the following are provisioned:
- **Cloud SQL (PostgreSQL)**: Version 15+
- **Cloud Memorystore (Redis)**: Version 7.0+
- **VPC Access Connector**: Required for Cloud Run to reach private IP instances.
- **Artifact Registry**: A repository named `exotics-lanka-repo`.

## 3. Secrets

The monolith requires the following secrets in **Secret Manager**:
- `DATABASE_URL`
- `REDIS_URL`
- `JWT_SECRET`
- `CLOUDINARY_URL` (optional)

## 4. Deployment

Deploy using Google Cloud Build:
```bash
gcloud builds submit --config cloudbuild.yaml .
```

This will:
1. Build the Docker image using the root `Dockerfile`.
2. Push the image to Artifact Registry.
3. Deploy to Cloud Run with the necessary secrets and VPC connector.

---
> [!NOTE]
> All configurations (Region, Repository Name, Service Name) can be adjusted in the `substitutions` section of [cloudbuild.yaml](../cloudbuild.yaml).
