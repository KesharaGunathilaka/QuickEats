#!/usr/bin/env bash
# Deploy one image tag of the QuickEats stack: pull, migrate, start.
# Usage: deploy.sh <image-tag>
# Used by Jenkins for both normal deploys and rollbacks.
set -euo pipefail

TAG="${1:?usage: deploy.sh <image-tag>}"
cd "${DEPLOY_DIR:-/opt/quickeats}"

# Shell env overrides IMAGE_TAG from .env.prod
export IMAGE_TAG="$TAG"
dc() { docker compose -p quickeats -f docker-compose.prod.yml --env-file .env.prod "$@"; }

echo "==> Pulling images for $TAG"
dc pull --quiet identity catalog order payment notification

echo "==> Running migrations"
dc run --rm migrate

echo "==> Starting stack on $TAG"
dc up -d --remove-orphans
