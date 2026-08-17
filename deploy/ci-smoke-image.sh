#!/usr/bin/env bash
set -euo pipefail

image=${1:?image tag is required}
expected_arch=${2:?expected architecture is required}
suffix="${GITHUB_RUN_ID:-local}-${GITHUB_RUN_ATTEMPT:-1}-${expected_arch}"
network="sub2api-smoke-${suffix}"
postgres="sub2api-smoke-postgres-${suffix}"
redis="sub2api-smoke-redis-${suffix}"
app="sub2api-smoke-app-${suffix}"

cleanup() {
  docker rm -f "$app" "$redis" "$postgres" >/dev/null 2>&1 || true
  docker network rm "$network" >/dev/null 2>&1 || true
}
trap cleanup EXIT

cleanup
docker network create "$network" >/dev/null

docker run -d --name "$postgres" --network "$network" \
  -e POSTGRES_USER=sub2api \
  -e POSTGRES_PASSWORD=ci_sub2api_password \
  -e POSTGRES_DB=sub2api \
  postgres:17-alpine >/dev/null

docker run -d --name "$redis" --network "$network" redis:7-alpine >/dev/null

for attempt in $(seq 1 60); do
  if docker exec "$postgres" pg_isready -U sub2api -d sub2api >/dev/null 2>&1 && \
    docker exec "$redis" redis-cli ping 2>/dev/null | grep -qx PONG; then
    break
  fi
  if [[ "$attempt" == 60 ]]; then
    docker logs "$postgres" || true
    docker logs "$redis" || true
    exit 1
  fi
  sleep 2
done

docker run -d --name "$app" --network "$network" --read-only \
  --tmpfs /app/data:rw,nosuid,size=64m \
  --tmpfs /tmp:rw,noexec,nosuid,size=64m \
  -e AUTO_SETUP=true \
  -e SERVER_HOST=0.0.0.0 \
  -e SERVER_PORT=8080 \
  -e SERVER_MODE=release \
  -e RUN_MODE=standard \
  -e DATABASE_HOST="$postgres" \
  -e DATABASE_PORT=5432 \
  -e DATABASE_USER=sub2api \
  -e DATABASE_PASSWORD=ci_sub2api_password \
  -e DATABASE_DBNAME=sub2api \
  -e DATABASE_SSLMODE=disable \
  -e REDIS_HOST="$redis" \
  -e REDIS_PORT=6379 \
  -e REDIS_DB=0 \
  -e ADMIN_EMAIL=ci-admin@example.invalid \
  -e 'ADMIN_PASSWORD=CiOnly123!ChangeMe' \
  -e JWT_SECRET=0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef \
  -e TOTP_ENCRYPTION_KEY=abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789 \
  "$image" >/dev/null

for attempt in $(seq 1 90); do
  status=$(docker inspect "$app" --format '{{.State.Status}}')
  if [[ "$status" == exited || "$status" == dead ]]; then
    docker logs "$app"
    exit 1
  fi
  if docker exec "$app" wget -q -T 5 -O /tmp/health.json http://127.0.0.1:8080/health; then
    break
  fi
  if [[ "$attempt" == 90 ]]; then
    docker logs "$app"
    exit 1
  fi
  sleep 2
done

test "$(docker image inspect "$image" --format '{{.Architecture}}')" = "$expected_arch"
docker exec "$app" sh -c "awk '/^Uid:/{exit !(\$2 == 1000)}' /proc/1/status"
docker exec "$app" sh -c "grep -q '\"ok\":true' /tmp/health.json"
docker exec --user 1000 "$app" /app/sub2api --version
