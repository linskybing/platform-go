#!/bin/bash

# Build Go API image
echo "Building Go API image..."
docker build -t platform-go-api:latest .

# Tag and push Go API image
echo "Tagging and pushing Go API image..."
docker tag platform-go-api:latest linskybing/platform-go-api:latest
docker push linskybing/platform-go-api:latest

# Build Postgres image
echo "Building Postgres image..."
docker build -t postgres-with-pg_cron:latest -f infra/db/postgres/Dockerfile infra/db/postgres

# Tag and push Postgres image
echo "Tagging and pushing Postgres image..."
docker tag postgres-with-pg_cron:latest linskybing/postgres-with-pg_cron:latest
docker push linskybing/postgres-with-pg_cron:latest

echo "Images built and pushed successfully."
