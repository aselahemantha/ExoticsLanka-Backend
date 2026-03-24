#!/bin/bash

# Configuration
PROJECT_ID="your-project-id"
REGION="us-central1"
ARTIFACT_REGISTRY="exotics-lanka-repo"

# Service Names
AUTH_SERVICE="auth-service"
LISTINGS_SERVICE="listings-service"
IMAGE_SERVICE="image-service"

# Build and Push Images
build_and_push() {
  local service=$1
  local dockerfile_path="services/$service/Dockerfile"
  local image_tag="$REGION-docker.pkg.dev/$PROJECT_ID/$ARTIFACT_REGISTRY/$service:latest"

  echo "Building $service..."
  docker build -t "$image_tag" -f "$dockerfile_path" "services/$service"
  
  echo "Pushing $service..."
  docker push "$image_tag"
}

# Deploy to Cloud Run
deploy_to_cloud_run() {
  local service=$1
  local image_tag="$REGION-docker.pkg.dev/$PROJECT_ID/$ARTIFACT_REGISTRY/$service:latest"

  echo "Deploying $service to Cloud Run..."
  gcloud run deploy "$service" \
    --image "$image_tag" \
    --region "$REGION" \
    --platform managed \
    --allow-unauthenticated \
    --vpc-connector "projects/$PROJECT_ID/locations/$REGION/connectors/exotics-lanka-connector" \
    --set-secrets="DATABASE_URL=DATABASE_URL:latest,REDIS_URL=REDIS_URL:latest,JWT_SECRET=JWT_SECRET:latest,CLOUDINARY_URL=CLOUDINARY_URL:latest"
}

# Main Execution
# build_and_push $AUTH_SERVICE
# build_and_push $LISTINGS_SERVICE
# build_and_push $IMAGE_SERVICE

# deploy_to_cloud_run $AUTH_SERVICE
# deploy_to_cloud_run $LISTINGS_SERVICE
# deploy_to_cloud_run $IMAGE_SERVICE

echo "Deployment completed!"
