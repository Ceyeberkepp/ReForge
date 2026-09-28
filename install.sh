#!/usr/bin/env bash
set -Eeuo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run ReForge installer as root." >&2
  exit 1
fi

command -v docker >/dev/null 2>&1 || {
  echo "Docker is required. Install Docker Engine and the Compose plugin first." >&2
  exit 1
}
docker compose version >/dev/null 2>&1 || {
  echo "Docker Compose plugin is required." >&2
  exit 1
}

ROOT="${REFORGE_ROOT:-/opt/reforge}"
mkdir -p "${ROOT}"
if [[ "${PWD}" != "${ROOT}" ]]; then
  cp -a . "${ROOT}/"
fi
cd "${ROOT}"

if [[ ! -f .env ]]; then
  cp .env.example .env
  DBPASS="$(python3 - <<'PY'
import secrets
print(secrets.token_urlsafe(24))
PY
)"
  sed -i "s/POSTGRES_PASSWORD=change-me/POSTGRES_PASSWORD=${DBPASS}/" .env
  sed -i "s#reforge:change-me@db#reforge:${DBPASS}@db#" .env
fi

docker compose -f infra/docker-compose.yml up -d --build

echo
echo "ReForge is starting."
echo "Web UI: http://$(hostname -I | awk '{print $1}'):5173"
echo "API:    http://$(hostname -I | awk '{print $1}'):8080"
echo "Docs:   http://$(hostname -I | awk '{print $1}'):8080/docs"
