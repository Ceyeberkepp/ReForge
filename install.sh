#!/usr/bin/env bash
set -Eeuo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run ReForge installer as root." >&2
  exit 1
fi

if [[ ! -r /etc/os-release ]]; then
  echo "ReForge one-click install currently supports Debian and Ubuntu Linux." >&2
  exit 1
fi

. /etc/os-release

case "${ID:-}" in
  debian|ubuntu) ;;
  *)
    echo "Unsupported distribution: ${PRETTY_NAME:-${ID:-unknown}}" >&2
    echo "Use Debian or Ubuntu for the one-click installer." >&2
    exit 1
    ;;
esac

# Validate installation safety before modifying host packages or services.
ROOT="${REFORGE_ROOT:-/opt/reforge}"
# Refuse unsafe or incomplete existing installations BEFORE changing deployed code.
# In particular, do not follow a symlinked credential file during upgrade.
if [[ -z "${ROOT}" || "${ROOT}" == "/" || "${ROOT}" == "." || "${ROOT}" == ".." ]]; then
  echo "REFORGE_ROOT must identify a dedicated installation directory." >&2
  exit 1
fi
# Refuse redirects of the deployment directory itself. A symlink could
# otherwise redirect code extraction outside the intended install tree.
if [[ "${ROOT}" != /* ]]; then
  echo "REFORGE_ROOT must be an absolute path." >&2
  exit 1
fi
# Resolve existing parent components as well as the final directory. This
# prevents /safe/link-to-elsewhere/reforge and /opt/../etc-style redirects.
# realpath -m accepts a not-yet-created final installation directory.
CANONICAL_ROOT="$(realpath -m -- "${ROOT}")"
if [[ "${ROOT}" != "${CANONICAL_ROOT}" ]]; then
  echo "REFORGE_ROOT must be a canonical absolute path without symlinked parents or traversal." >&2
  exit 1
fi
if [[ -L "${ROOT}" ]]; then
  echo "Refusing a symlinked ReForge installation root." >&2
  exit 1
fi
# Do not copy a local checkout into an unrelated populated destination.
if [[ -f "./infra/docker-compose.yml" && "$(pwd -P)" != "${ROOT}" &&
      -d "${ROOT}" && ! -f "${ROOT}/infra/docker-compose.yml" && ! -f "${ROOT}/.env" &&
      -n "$(ls -A "${ROOT}" 2>/dev/null || true)" ]]; then
  echo "REFORGE_ROOT contains unrelated files; refusing to overwrite them." >&2
  exit 1
fi
# Refuse FIFOs, sockets, devices and directories before any credential reads.
# A FIFO can otherwise hang the installer during grep-based preflight.
if [[ -e "${ROOT}/.env" && ! -f "${ROOT}/.env" ]]; then
  echo "Existing .env must be a regular credential file; refusing unsafe upgrade." >&2
  exit 1
fi
if [[ -L "${ROOT}/.env" ]]; then
  echo "Refusing to use a symlinked production .env file." >&2
  exit 1
fi
# A hardlinked .env shares an inode with another path. Appending an upgrade
# setting would silently change that other file, so fail before host mutation.
if [[ -f "${ROOT}/.env" ]] && (( $(stat -c %h -- "${ROOT}/.env") != 1 )); then
  echo "Existing .env must not be hardlinked; refusing unsafe upgrade." >&2
  exit 1
fi
# Production credentials must not be readable by group or other users.
# Refuse unsafe permissions before copying code or invoking package managers;
# do not silently chmod a pre-existing production file.
if [[ -f "${ROOT}/.env" ]]; then
  ENV_MODE="$(stat -c %a -- "${ROOT}/.env")"
  if (( (8#${ENV_MODE} & 077) != 0 )); then
    echo "Existing .env permissions expose credentials to group or others; set mode 600 before upgrading." >&2
    exit 1
  fi
fi
# A previously deployed Compose stack without its credential file must not be
# treated as a new install. That would silently generate different database
# credentials against persistent volumes and strand existing data.
if [[ -f "${ROOT}/infra/docker-compose.yml" && ! -f "${ROOT}/.env" ]]; then
  echo "Existing ReForge deployment is missing .env; restore its credentials from backup before upgrading." >&2
  exit 1
fi
if [[ -e "${ROOT}/.env" ]]; then
  # Different dotenv readers may resolve repeated credentials differently.
  # Reject ambiguity before packages, services, or deployed files change.
  for SECRET_KEY in REFORGE_WORKER_TOKEN REFORGE_DB_DSN; do
    SECRET_COUNT="$(grep -c "^${SECRET_KEY}=" "${ROOT}/.env" || true)"
    if [[ "${SECRET_COUNT}" != 1 ]]; then
      echo "Existing .env must contain exactly one ${SECRET_KEY}; repair before upgrading." >&2
      exit 1
    fi
  done
  if [[ ! -f "${ROOT}/.env" ]] ||
     ! grep -q '^REFORGE_WORKER_TOKEN=.' "${ROOT}/.env" ||
     ! grep -q '^REFORGE_DB_DSN=.' "${ROOT}/.env"; then
    echo "Existing .env is invalid; refusing to modify installed code." >&2
    exit 1
  fi
fi
# Validate the persisted port on upgrades before changing host packages.
# An explicit shell override is intentionally ignored for existing installs.
if [[ -f "${ROOT}/.env" ]] && grep -q '^REFORGE_WEB_PORT=' "${ROOT}/.env"; then
  # Duplicate keys are ambiguous across dotenv/Compose readers; fail closed.
  PORT_KEY_COUNT="$(grep -c '^REFORGE_WEB_PORT=' "${ROOT}/.env")"
  if (( PORT_KEY_COUNT != 1 )); then
    echo "Existing .env has duplicate REFORGE_WEB_PORT entries; repair before upgrading." >&2
    exit 1
  fi
  SAVED_WEB_PORT="$(sed -n 's/^REFORGE_WEB_PORT=//p' "${ROOT}/.env" | tail -n 1)"
  if [[ ! "${SAVED_WEB_PORT}" =~ ^[1-9][0-9]{3,4}$ ]] ||
     (( 10#${SAVED_WEB_PORT:-0} < 1024 || 10#${SAVED_WEB_PORT:-0} > 65535 )); then
    echo "Existing REFORGE_WEB_PORT is invalid; repair .env before upgrading." >&2
    exit 1
  fi
fi
# Reject invalid explicit web ports before changing deployment files. An upgrade
# keeps its persisted port; fresh installs may request a different starting port.
if [[ ! -f "${ROOT}/.env" && -n "${REFORGE_WEB_PORT:-}" ]]; then
  if [[ ! "${REFORGE_WEB_PORT}" =~ ^[1-9][0-9]{3,4}$ ]] ||
     (( 10#${REFORGE_WEB_PORT} < 1024 || 10#${REFORGE_WEB_PORT} > 65535 )); then
    echo "REFORGE_WEB_PORT must be an integer from 1024 to 65535." >&2
    exit 1
  fi
fi
# Validate explicit fresh-install choices before copying or cloning deployed code.
# Existing installations retain their persisted database and service configuration.
if [[ ! -f "${ROOT}/.env" ]]; then
  case "${REFORGE_DB_CHOICE:-}" in
    ""|1|2|3) ;;
    *) echo "Invalid database selection (expected 1, 2, or 3)." >&2; exit 1 ;;
  esac
  for SERVICE_VALUE in "${REFORGE_ENABLE_MONGODB:-}" "${REFORGE_ENABLE_NEO4J:-}"; do
    case "${SERVICE_VALUE,,}" in
      ""|y|yes|1|n|no|0) ;;
      *) echo "Optional service selections must be yes/no, y/n, or 1/0." >&2; exit 1 ;;
    esac
  done
fi

echo "[1/6] Installing base system dependencies..."
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y --no-install-recommends \
  ca-certificates \
  curl \
  sudo \
  git \
  openssl \
  gnupg \
  coreutils \
  tar \
  iproute2

if ! command -v docker >/dev/null 2>&1 || ! docker compose version >/dev/null 2>&1; then
  echo "[2/6] Installing Docker Engine and Docker Compose..."

  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL "https://download.docker.com/linux/${ID}/gpg" \
    | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
  chmod a+r /etc/apt/keyrings/docker.gpg

  ARCH="$(dpkg --print-architecture)"
  CODENAME="${VERSION_CODENAME:-}"
  if [[ -z "${CODENAME}" ]]; then
    CODENAME="$(. /etc/os-release && echo "${UBUNTU_CODENAME:-}")"
  fi
  if [[ -z "${CODENAME}" ]]; then
    echo "Unable to determine Debian/Ubuntu release codename." >&2
    exit 1
  fi

  cat >/etc/apt/sources.list.d/docker.sources <<EOF
Types: deb
URIs: https://download.docker.com/linux/${ID}
Suites: ${CODENAME}
Components: stable
Architectures: ${ARCH}
Signed-By: /etc/apt/keyrings/docker.gpg
EOF

  apt-get update
  apt-get install -y --no-install-recommends \
    docker-ce \
    docker-ce-cli \
    containerd.io \
    docker-buildx-plugin \
    docker-compose-plugin
else
  echo "[2/6] Docker Engine and Compose are already installed."
fi

echo "[3/6] Starting Docker..."
systemctl enable --now docker 2>/dev/null || true

if ! docker info >/dev/null 2>&1; then
  if command -v service >/dev/null 2>&1; then
    service docker start 2>/dev/null || true
  fi
fi

if ! docker info >/dev/null 2>&1; then
  echo "Docker is installed but the daemon is not available." >&2
  echo "If this is an LXC/container, enable nesting/container-runtime support and restart the guest." >&2
  exit 1
fi

docker compose version >/dev/null 2>&1 || {
  echo "Docker Compose plugin installation failed." >&2
  exit 1
}

secret() {
  openssl rand -hex 24
}

if [[ -f "./infra/docker-compose.yml" ]]; then
  mkdir -p "${ROOT}"
  if [[ "$(pwd)" != "${ROOT}" ]]; then
    # Never overwrite a deployed .env with a developer checkout's .env.
    # Copy code and assets while preserving runtime credentials and local Git metadata.
    tar --exclude='./.env' --exclude='./.git' -cf - . | tar -C "${ROOT}" -xf -
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
  # Git can clone into an existing empty directory. Never recursively delete
  # an installation target, even when it appeared empty during preflight.
  git clone --depth 1 https://github.com/Ceyeberkepp/ReForge.git "${ROOT}"
fi
cd "${ROOT}"

HOST_IP="${REFORGE_HOST_IP:-$(hostname -I 2>/dev/null | awk '{print $1}')}"
HOST_IP="${HOST_IP:-127.0.0.1}"

echo
echo "[4/6] Configuring ReForge..."
# Existing installations must NEVER rotate generated secrets or database passwords.
if [[ -f .env ]]; then
  echo "Existing ReForge installation detected; preserving .env credentials."
  if ! grep -q '^REFORGE_WORKER_TOKEN=' .env || ! grep -q '^REFORGE_DB_DSN=' .env; then
    echo "Existing .env is incomplete; refusing to overwrite it." >&2
    exit 1
  fi
  if ! grep -q '^REFORGE_WEB_PORT=' .env; then
    echo "REFORGE_WEB_PORT=5173" >>.env
  fi
  ADMINPASS=""
else
  # On fresh installs select an unused host port; preserve the choice on upgrades.
  WEB_PORT="${REFORGE_WEB_PORT:-5173}"
  if ! [[ "${WEB_PORT}" =~ ^[0-9]+$ ]] || (( WEB_PORT < 1024 || WEB_PORT > 65535 )); then
    echo "REFORGE_WEB_PORT must be an integer from 1024 to 65535." >&2
    exit 1
  fi
  while ss -H -ltn "sport = :${WEB_PORT}" 2>/dev/null | grep -q .; do
    (( WEB_PORT < 65535 )) || { echo "No free ReForge web port found." >&2; exit 1; }
    WEB_PORT=$((WEB_PORT + 1))
  done
  echo "Selected web port: ${WEB_PORT}"
echo
echo "ReForge database"
echo "  1) PostgreSQL (recommended)"
echo "  2) MariaDB / MySQL"
echo "  3) Microsoft SQL Server Express"
DB_CHOICE="${REFORGE_DB_CHOICE:-}"
if [[ -z "${DB_CHOICE}" ]]; then
  if [[ -n "${REFORGE_NONINTERACTIVE:-}" || ! -t 0 ]]; then
    DB_CHOICE=1
  else
    read -r -p "Choose [1]: " DB_CHOICE
    DB_CHOICE="${DB_CHOICE:-1}"
  fi
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
if [[ -z "${REFORGE_NONINTERACTIVE:-}" && -t 0 ]]; then
  if [[ -z "${ENABLE_MONGO}" ]]; then
    read -r -p "Enable optional MongoDB service? [y/N]: " ENABLE_MONGO || true
  fi
  if [[ -z "${ENABLE_NEO4J}" ]]; then
    read -r -p "Enable optional Neo4j graph service? [y/N]: " ENABLE_NEO4J || true
  fi
fi
# Reject mistyped explicit choices rather than silently disabling a requested service.
# Accept the documented yes/no forms; omitted values remain disabled by default.
for SERVICE_VALUE in "${ENABLE_MONGO}" "${ENABLE_NEO4J}"; do
  case "${SERVICE_VALUE,,}" in
    ""|y|yes|1|n|no|0) ;;
    *)
      echo "Optional service selections must be yes/no, y/n, or 1/0." >&2
      exit 1
      ;;
  esac
done
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
REFORGE_ALLOWED_ORIGIN=http://${HOST_IP}:${WEB_PORT}
REFORGE_WEB_PORT=${WEB_PORT}
VITE_API_URL=
EOF
chmod 600 .env

if [[ "${DB_CHOICE}" == "3" ]]; then
  echo "Starting Microsoft SQL Server..."
  docker compose --env-file "${ROOT}/.env" -f infra/docker-compose.yml --profile sqlserver up -d sqlserver
  echo "Waiting for SQL Server..."
  SQLSERVER_READY=false
  for _ in $(seq 1 45); do
    if docker compose --env-file "${ROOT}/.env" -f infra/docker-compose.yml exec -T sqlserver bash -lc '
      SQLCMD=""
      [[ -x /opt/mssql-tools18/bin/sqlcmd ]] && SQLCMD=/opt/mssql-tools18/bin/sqlcmd
      [[ -z "$SQLCMD" && -x /opt/mssql-tools/bin/sqlcmd ]] && SQLCMD=/opt/mssql-tools/bin/sqlcmd
      [[ -n "$SQLCMD" ]] && "$SQLCMD" -C -S localhost -U sa -P "$MSSQL_SA_PASSWORD" -Q "IF DB_ID('"'"'reforge'"'"') IS NULL CREATE DATABASE reforge" >/dev/null
    '; then
      SQLSERVER_READY=true
      break
    fi
    sleep 2
  done
  if [[ "$SQLSERVER_READY" != true ]]; then
    echo "SQL Server database initialization did not succeed; refusing to report a successful installation." >&2
    exit 1
  fi
fi
fi

echo "[5/6] Building and starting ReForge..."
docker compose --env-file "${ROOT}/.env" -f infra/docker-compose.yml up -d --build

echo "[6/6] Verifying ReForge services..."
docker compose --env-file "${ROOT}/.env" -f infra/docker-compose.yml ps
# Container start does not establish that the API and database are usable.
# Test the published web endpoint (which proxies /health) with a bounded retry.
WEB_PORT="$(sed -n 's/^REFORGE_WEB_PORT=//p' .env | tail -n 1)"
WEB_PORT="${WEB_PORT:-5173}"
HEALTH_READY=false
for _ in $(seq 1 30); do
  if curl --noproxy '*' --fail --silent --show-error --max-time 3 \
      "http://127.0.0.1:${WEB_PORT}/health" 2>/dev/null | grep -q '"status":"ok"'; then
    HEALTH_READY=true
    break
  fi
  sleep 2
done
if [[ "$HEALTH_READY" != true ]]; then
  echo "ReForge API did not become healthy through the web endpoint; refusing to report successful installation." >&2
  exit 1
fi

echo
echo "ReForge API health verified through the web endpoint."
WEB_PORT="$(sed -n 's/^REFORGE_WEB_PORT=//p' .env | tail -n 1)"
WEB_PORT="${WEB_PORT:-5173}"
echo "Web UI: http://${HOST_IP}:${WEB_PORT}"
echo
if [[ -n "${ADMINPASS}" ]]; then
  echo "Initial administrator:"
  echo "  Username: admin"
  echo "  Password: ${ADMINPASS}"
  echo "Save the password now. Change it after first sign-in."
else
  echo "Existing administrator credentials were preserved (not printed or reset)."
fi
echo
echo "PXE configuration is available inside the web UI under PXE / Network."
