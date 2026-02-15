#!/bin/bash

# Usage: ./scripts/deploy.sh <service_name>
# Example: ./scripts/deploy.sh auth-service

SERVICE_NAME=$1
PROJECT_ID=$(gcloud config get-value project)
REGION="us-central1"

if [ -z "$SERVICE_NAME" ]; then
  echo "Usage: $0 <service_name>"
  exit 1
fi

if [ -z "$PROJECT_ID" ]; then
  echo "Error: No Google Cloud Project ID found. Please run 'gcloud config set project <PROJECT_ID>'."
  exit 1
fi

echo "Deploying $SERVICE_NAME to project $PROJECT_ID in region $REGION..."

SERVICE_DIR="services/$SERVICE_NAME"

if [ ! -d "$SERVICE_DIR" ]; then
  echo "Error: Service directory $SERVICE_DIR does not exist."
  exit 1
fi

# Determine Secrets and Cloud SQL Connection
# We check if specific secrets exist and map them if they do
SECRETS=""
CLOUD_SQL_INSTANCES=""

# Check for DB Secret
if gcloud secrets describe exoticslanka-db-url --project="$PROJECT_ID" > /dev/null 2>&1; then
  echo "Found secret: exoticslanka-db-url. Mapping to DATABASE_URL."
  SECRETS="$SECRETS,DATABASE_URL=exoticslanka-db-url:latest"
  
  # Try to extract instance connection name from secret or convention
  # For simplicity, we assume a standard name or require it to be set in env
  # We'll try to look it up from the project
  INSTANCE_NAME=$(gcloud sql instances list --format="value(connectionName)" --filter="name:exoticslanka*" --limit=1)
  if [ -n "$INSTANCE_NAME" ]; then
      echo "Found Cloud SQL Instance: $INSTANCE_NAME"
      CLOUD_SQL_INSTANCES="--add-cloudsql-instances=$INSTANCE_NAME"
  fi
fi

# Check for Redis Secret
if gcloud secrets describe exoticslanka-redis-url --project="$PROJECT_ID" > /dev/null 2>&1; then
  echo "Found secret: exoticslanka-redis-url. Mapping to REDIS_URL."
  # Add comma if SECRETS is not empty
  if [ -n "$SECRETS" ]; then SECRETS="$SECRETS,"; fi
  SECRETS="${SECRETS}REDIS_URL=exoticslanka-redis-url:latest"
fi

# Clean up leading comma if explicitly set incorrectly (safety)
SECRETS=$(echo "$SECRETS" | sed 's/^,//')

cd "$SERVICE_DIR" || exit

# 1. Build and Push Image using Google Cloud Build
echo "Building and pushing image for $SERVICE_NAME..."
gcloud builds submit --tag "gcr.io/$PROJECT_ID/$SERVICE_NAME" .

# 2. Deployment Command Builder
DEPLOY_CMD="gcloud run deploy $SERVICE_NAME \
  --image gcr.io/$PROJECT_ID/$SERVICE_NAME \
  --platform managed \
  --region $REGION \
  --allow-unauthenticated \
  --set-env-vars PROJECT_ID=$PROJECT_ID"

if [ -n "$SECRETS" ]; then
    DEPLOY_CMD="$DEPLOY_CMD --set-secrets=$SECRETS"
fi

if [ -n "$CLOUD_SQL_INSTANCES" ]; then
    DEPLOY_CMD="$DEPLOY_CMD $CLOUD_SQL_INSTANCES"
fi

# Execute Deployment
echo "Deploying to Cloud Run with command:"
echo "$DEPLOY_CMD"
eval "$DEPLOY_CMD"

echo "Deployment of $SERVICE_NAME complete!"
