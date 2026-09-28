#!/usr/bin/env bash
set -Eeuo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run ReForge installer as root." >&2
  exit 1
fi

command -v docker >/dev/null 2>&1 || {
  echo "Docker Engine is required." >&2
  exit 1
}
command -v openssl >/dev/null 2>&1 || {
  echo "OpenSSL is required to generate installation secrets." >&2
  exit 1
}
docker compose version >/dev/null 2>&1 || {
  echo "Docker Compose plugin is required." >&2
  exit 1
}

secret() {
  openssl rand -hex 24
}

ROOT="${REFORGE_ROOT:-/opt/reforge}"
if [[ -f "./infra/docker-compose.yml" ]]; then
  mkdir -p "${ROOT}"
  if [[ "$(pwd)" != "${ROOT}" ]]; then
    cp -a . "${ROOT}/"
  fi
elif [[ ! -f "${ROOT}/infra/docker-compose.yml" ]]; then
  command -v git >/dev/null 2>&1 || {
    echo "Git is required for one-command installation." >&2
    exit 1
  }
  if [[ -e "${ROOT}" && -n "$(ls -A "${ROOT}" 2>/dev/null || true)" ]]; then
    echo "${ROOT} exists and is not a ReForge installation." >&2
    exit 1
  fi
  rm -rf "${ROOT}"
  git clone --depth 1 https://github.com/Ceyeberkepp/ReForge.git "${ROOT}"
fi
cd "${ROOT}"

HOST_IP="${REFORGE_HOST_IP:-$(hostname -I 2>/dev/null | awk '{print $1}')}"
HOST_IP="${HOST_IP:-127.0.0.1}"

echo
echo "ReForge database"
echo "  1) PostgreSQL (recommended)"
echo "  2) MariaDB / MySQL"
echo "  3) Microsoft SQL Server Express"
DB_CHOICE="${REFORGE_DB_CHOICE:-}"
if [[ -z "${DB_CHOICE}" ]]; then
  read -r -p "Choose [1]: " DB_CHOICE
  DB_CHOICE="${DB_CHOICE:-1}"
fi

DBPASS="$(secret 36)"
ROOTPASS="$(secret 40)"
ADMINPASS="${REFORGE_ADMIN_PASSWORD:-}"
if [[ -z "${ADMINPASS}" ]]; then
  ADMINPASS="Rf!$(secret 28)9a"
fi
WORKERTOKEN="$(secret 64)"
MONGOPASS="$(secret 40)"
NEO4JPASS="Rf!$(secret 30)9a"

PROFILE=""
DB_DRIVER=""
DB_DSN=""

case "${DB_CHOICE}" in
  1)
    PROFILE="postgres"
    DB_DRIVER="postgres"
    DB_DSN="host=postgres user=reforge password=${DBPASS} dbname=reforge port=5432 sslmode=disable"
    ;;
  2)
    PROFILE="mariadb"
    DB_DRIVER="mariadb"
    DB_DSN="reforge:${DBPASS}@tcp(mariadb:3306)/reforge?charset=utf8mb4&parseTime=True&loc=UTC"
    ;;
  3)
    PROFILE="sqlserver"
    DB_DRIVER="sqlserver"
    MSSQLPASS="Rf!$(secret 28)9a"
    DBPASS="${MSSQLPASS}"
    DB_DSN="sqlserver://sa:${MSSQLPASS}@sqlserver:1433?database=reforge&encrypt=disable"
    ;;
  *)
    echo "Invalid database selection." >&2
    exit 1
    ;;
esac

ENABLE_MONGO="${REFORGE_ENABLE_MONGODB:-}"
ENABLE_NEO4J="${REFORGE_ENABLE_NEO4J:-}"
if [[ -z "${REFORGE_NONINTERACTIVE:-}" ]]; then
  read -r -p "Enable optional MongoDB service? [y/N]: " ENABLE_MONGO || true
  read -r -p "Enable optional Neo4j graph service? [y/N]: " ENABLE_NEO4J || true
fi
[[ "${ENABLE_MONGO,,}" == "y" || "${ENABLE_MONGO,,}" == "yes" || "${ENABLE_MONGO}" == "1" ]] && PROFILE+=",mongodb"
[[ "${ENABLE_NEO4J,,}" == "y" || "${ENABLE_NEO4J,,}" == "yes" || "${ENABLE_NEO4J}" == "1" ]] && PROFILE+=",neo4j"

cat > .env <<EOF
COMPOSE_PROFILES=${PROFILE}
REFORGE_DB_DRIVER=${DB_DRIVER}
REFORGE_DB_DSN=${DB_DSN}

POSTGRES_DB=reforge
POSTGRES_USER=reforge
POSTGRES_PASSWORD=${DBPASS}

MYSQL_DATABASE=reforge
MYSQL_USER=reforge
MYSQL_PASSWORD=${DBPASS}
MYSQL_ROOT_PASSWORD=${ROOTPASS}

MSSQL_SA_PASSWORD=${DBPASS}

MONGO_ROOT_USER=reforge
MONGO_ROOT_PASSWORD=${MONGOPASS}
NEO4J_AUTH=neo4j/${NEO4JPASS}

REFORGE_ADMIN_USER=admin
REFORGE_ADMIN_PASSWORD=${ADMINPASS}
REFORGE_WORKER_TOKEN=${WORKERTOKEN}
REFORGE_ALLOWED_ORIGIN=http://${HOST_IP}:5173
VITE_API_URL=
EOF
chmod 600 .env

if [[ "${DB_CHOICE}" == "3" ]]; then
  echo "Starting Microsoft SQL Server..."
  docker compose -f infra/docker-compose.yml --profile sqlserver up -d sqlserver
  echo "Waiting for SQL Server..."
  for _ in $(seq 1 45); do
    if docker compose -f infra/docker-compose.yml exec -T sqlserver bash -lc '
      SQLCMD=""
      [[ -x /opt/mssql-tools18/bin/sqlcmd ]] && SQLCMD=/opt/mssql-tools18/bin/sqlcmd
      [[ -z "$SQLCMD" && -x /opt/mssql-tools/bin/sqlcmd ]] && SQLCMD=/opt/mssql-tools/bin/sqlcmd
      [[ -n "$SQLCMD" ]] && "$SQLCMD" -C -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -Q "IF DB_ID('"'"'reforge'"'"') IS NULL CREATE DATABASE reforge" >/dev/null
    '; then
      break
    fi
    sleep 2
  done
fi

docker compose -f infra/docker-compose.yml up -d --build

echo
echo "ReForge installation started successfully."
echo "Web UI: http://${HOST_IP}:5173"
echo
echo "Initial administrator:"
echo "  Username: admin"
echo "  Password: ${ADMINPASS}"
echo
echo "Save the password now. Change it after first sign-in."
echo
echo "PXE configuration is available inside the web UI under PXE / Network."
