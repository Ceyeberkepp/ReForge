# KB-001 — Install and update ReForge

## Install

Supported target: Docker-capable Debian/Ubuntu system.

```bash
curl -fsSL https://raw.githubusercontent.com/Ceyeberkepp/ReForge/main/install.sh | sudo bash
```

The installer installs required host packages, Docker Engine/Compose when needed, clones ReForge to `/opt/reforge`, asks for the database profile, generates secrets, and starts the stack.

## Update

```bash
cd /opt/reforge
git pull

docker compose --env-file /opt/reforge/.env \
  -f /opt/reforge/infra/docker-compose.yml \
  build api worker web

docker compose --env-file /opt/reforge/.env \
  -f /opt/reforge/infra/docker-compose.yml \
  up -d
```

## Verify

```bash
docker compose --env-file /opt/reforge/.env \
  -f /opt/reforge/infra/docker-compose.yml ps

curl http://127.0.0.1:5173/health
```

The web service should publish port `5173`.
