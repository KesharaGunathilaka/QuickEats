#!/usr/bin/env bash
# One-time: copy data from the old shared Postgres into the per-service databases.
#
# Run on the server AFTER the first deploy that uses per-service databases (that deploy
# starts them empty, migrated and seeded). This replaces their contents with the data
# from the old shared volume, so users, orders, wallets etc. carry over.
#
#   sudo -u jenkins bash deploy/split-postgres.sh        (from /opt/quickeats)
#
# The old volume is only read, never changed or removed. Delete it yourself once the
# app works:  docker volume rm quickeats_pgdata
set -euo pipefail

cd "${DEPLOY_DIR:-/opt/quickeats}"
dc() { docker compose -p "${PROJECT:-quickeats}" -f docker-compose.prod.yml --env-file "${ENV_FILE:-.env.prod}" "$@"; }
OLD_VOLUME="${OLD_VOLUME:-quickeats_pgdata}"
OLD_CONTAINER="quickeats-old-postgres"
SERVICES="identity catalog order payment"

docker volume inspect "$OLD_VOLUME" >/dev/null 2>&1 || { echo "No volume $OLD_VOLUME: nothing to copy."; exit 0; }

cleanup() { docker rm -f "$OLD_CONTAINER" >/dev/null 2>&1 || true; }
trap cleanup EXIT

# Start the old data read-only in a throwaway container (no network exposure; the
# superuser password is irrelevant because we only use the local socket).
echo "==> Starting old shared Postgres from $OLD_VOLUME"
docker run -d --name "$OLD_CONTAINER" --network none \
  -v "$OLD_VOLUME":/var/lib/postgresql/data \
  -e POSTGRES_PASSWORD=unused postgres:16-alpine >/dev/null
until docker exec "$OLD_CONTAINER" pg_isready -U quickeats -q; do sleep 1; done

# Stop the services so nothing writes while their database is replaced.
echo "==> Stopping services"
dc stop $SERVICES

for svc in $SERVICES; do
  echo "==> Copying ${svc}_db"
  # Empty the new database (drops the seed and migration rows too; the dump has both).
  dc exec -T "$svc-db" psql -U "$svc" -d "${svc}_db" -v ON_ERROR_STOP=1 -q \
    -c 'DROP SCHEMA public CASCADE; CREATE SCHEMA public;'
  # --no-owner / --no-privileges: objects become owned by the new per-service user.
  docker exec "$OLD_CONTAINER" pg_dump -U quickeats --no-owner --no-privileges "${svc}_db" \
    | dc exec -T "$svc-db" psql -U "$svc" -d "${svc}_db" -v ON_ERROR_STOP=1 -q -o /dev/null
done

echo "==> Starting services"
dc start $SERVICES
echo "Done. Check the app, then remove the old volume: docker volume rm $OLD_VOLUME"
