#!/usr/bin/env bash
# Wait until every service container is healthy and Kong routes traffic to them.
# Gives up after ~2.5 minutes with a non-zero exit so Jenkins rolls back.
set -uo pipefail

cd "${DEPLOY_DIR:-/opt/quickeats}"
dc() { docker compose -p quickeats -f docker-compose.prod.yml --env-file .env.prod "$@"; }
SERVICES="identity catalog order payment notification"
ATTEMPTS=30

for attempt in $(seq 1 $ATTEMPTS); do
    failing=""

    # 1. Docker health status of each service (driven by GET /health)
    for s in $SERVICES; do
        id=$(dc ps -q "$s" 2>/dev/null)
        if [ -n "$id" ]; then
            status=$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "$id")
        else
            status="missing"
        fi
        [ "$status" = "healthy" ] || failing="$failing $s($status)"
    done

    # 2. End-to-end through Kong on port 80
    notif=$(curl -s -o /dev/null -w '%{http_code}' http://localhost/notification/health || true)
    [ "$notif" = "200" ] || failing="$failing kong->notification($notif)"

    # A login with no credentials must reach identity and come back 401
    auth=$(curl -s -o /dev/null -w '%{http_code}' -X POST -H 'Content-Type: application/json' \
        -d '{}' http://localhost/api/v1/auth/login || true)
    [ "$auth" = "401" ] || failing="$failing kong->identity($auth)"

    if [ -z "$failing" ]; then
        echo "All services healthy (IMAGE_TAG=${IMAGE_TAG:-unknown})"
        exit 0
    fi
    echo "[$attempt/$ATTEMPTS] waiting on:$failing"
    sleep 5
done

echo "Health check FAILED:$failing"
dc ps
exit 1
