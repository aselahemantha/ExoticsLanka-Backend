#!/bin/bash

# Script to deploy all services

SERVICES_DIR="services"

for SERVICE in "$SERVICES_DIR"/*; do
  if [ -d "$SERVICE" ]; then
    SERVICE_NAME=$(basename "$SERVICE")
    ./scripts/deploy.sh "$SERVICE_NAME"
  fi
done
