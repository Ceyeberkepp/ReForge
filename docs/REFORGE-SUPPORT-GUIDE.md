# ReForge Support Guide

This guide covers administration and first-line support for **ReForge**, the endpoint imaging, reimaging, PXE deployment, and device-management platform.

## 1. Architecture

ReForge currently consists of:

- **Go 1.27 API/control plane** — authentication, audit, configuration, PXE/iPXE generation, deployment orchestration, content metadata, users/RBAC
- **Go 1.27 worker** — deployment queue worker
- **React 19 + TypeScript web UI**
- **Nginx** — static web hosting and same-origin proxy
- **Transactional SQL database** — PostgreSQL, MariaDB/MySQL, or Microsoft SQL Server
- **Persistent deployment-content volume** — Gold Images, ISOs, driver packs, and application installers
- **Dedicated imaging-node boundary** — destructive disk capture/restore is not performed by the web/API container

The control plane can be fully operational even when no imaging node is attached. In that state, destructive capture/restore jobs must remain waiting rather than being reported as successful.

## 2. Default access

The default published web port is:

```text
5173
```

Typical address:

```text
http://SERVER-IP:5173
```

Health endpoint:

```text
/health
```

Example:

```bash
curl http://127.0.0.1:5173/health
```

A healthy response identifies the ReForge API, database driver, status, and version.

## 3. Main UI areas

ReForge currently exposes:

### Images
- Gold Images
- ISO Library
- Clone Images
- macOS Installers

### Deployment
- Hosts (PXE)
- Deployments
- Task Sequences
- Departments

### Management
- Applications
- Drivers
- Directory Services
- PXE / Network

### Monitoring
- Audit Logs
- Reports

### System
- Admin Center
- Settings

## 4. Deployment content

Uploaded content is stored persistently under the API content root, normally:

```text
/var/lib/reforge/content
```

Docker Compose persists it with the `reforge-content` volume.

Supported upload types currently include:

- Gold Images: `.wim`, `.esd`, `.img`, `.gz`, `.zst`, `.zip`
- ISO Library: `.iso`
- Driver Packs: `.zip`, `.cab`, `.7z`
- Applications: `.msi`, `.exe`, `.msix`, `.appx`, `.pkg`, `.zip`, `.ps1`, `.sh`

Uploads are streamed through Nginx to the API and stored outside the SQL database. SHA-256 is calculated for image/ISO/driver content.

## 5. PXE authentication flow

The deployment portal requires ReForge credentials.

Expected flow:

```text
PXE boot
  ↓
ReForge boot menu
  ↓
Install / Capture
  ↓
Username + password
  ↓
Permission check
  ↓
Deployment source menu
```

After authentication, permissions determine access to:

- ISO deployment
- Gold Image deployment
- Clone Image deployment
- image capture
- diagnostics

Local boot does not need to expose deployment content.

## 6. PXE branding and screen customization

PXE / Network supports customization for:

- Boot Menu
- Install Screen
- Capture Screen
- Loading/Progress Screen

Customizable fields include:

- organization/company name
- logo
- background image
- menu title/subtitle
- install title/subtitle
- capture title/subtitle
- loading title/message
- support/footer text
- accent color
- text color
- fallback panel/background color
- background overlay opacity
- default selection
- timeout
- which boot-menu entries are visible

The browser preview is intended to reflect the generated PXE experience, while actual firmware/iPXE graphics support may vary.

## 7. Host identity

A manually managed host does **not** require a MAC address.

ReForge can store:

- hostname
- serial number
- hardware UUID
- MAC address when available
- manufacturer/model
- department assignment

PXE flows may receive MAC and SMBIOS UUID from firmware. A PXE task can use hardware UUID when MAC is unavailable.

## 8. Users, groups and RBAC

ReForge supports local users, groups, and permissions.

Current permission families include:

- `pxe.login`
- `pxe.install`
- `pxe.capture`
- `pxe.iso`
- `pxe.gold`
- `pxe.clone`
- `pxe.diagnostics`
- `admin.users`
- `admin.groups`
- `admin.pxe`
- `admin.branding`

Use least privilege for operators who only need deployment access.

## 9. Directory Services

ReForge supports LDAP/Active Directory configuration and testing with:

- LDAPS
- StartTLS

Plain LDAP is disabled by default and should remain disabled unless an operator explicitly accepts that risk.

Directory configuration includes domain/DC information, Base DN, service account, default OU, and test credentials.

## 10. Database support

Supported primary database drivers:

- PostgreSQL
- MariaDB/MySQL
- Microsoft SQL Server

PostgreSQL is the recommended default.

Optional MongoDB and Neo4j services are not authoritative deployment stores.

## 11. Updating

Standard install path:

```text
/opt/reforge
```

Update:

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

Verify:

```bash
docker compose --env-file /opt/reforge/.env \
  -f /opt/reforge/infra/docker-compose.yml ps

curl http://127.0.0.1:5173/health
```

## 12. Logs

```bash
cd /opt/reforge

docker compose --env-file /opt/reforge/.env \
  -f infra/docker-compose.yml logs --tail=100 api

docker compose --env-file /opt/reforge/.env \
  -f infra/docker-compose.yml logs --tail=100 worker

docker compose --env-file /opt/reforge/.env \
  -f infra/docker-compose.yml logs --tail=100 web
```

## 13. Standard support triage

Collect:

```bash
cd /opt/reforge

git log -5 --oneline

docker compose --env-file /opt/reforge/.env \
  -f infra/docker-compose.yml ps

curl -sS http://127.0.0.1:5173/health

docker compose --env-file /opt/reforge/.env \
  -f infra/docker-compose.yml logs --tail=100 api

docker compose --env-file /opt/reforge/.env \
  -f infra/docker-compose.yml logs --tail=100 worker

docker compose --env-file /opt/reforge/.env \
  -f infra/docker-compose.yml logs --tail=100 web

ss -lntp | grep 5173 || true
df -h
docker volume ls
```

For PXE issues also collect:

- DHCP option/next-server configuration
- BIOS vs UEFI client type
- generated `/boot/ipxe` output
- whether the client reaches the login prompt
- whether the user has the required PXE permissions

## 14. Current imaging boundary

The ReForge control plane currently creates and authorizes deployment/capture tasks, but **raw disk imaging is intentionally isolated from the web/API containers**.

Until the dedicated imaging node is attached and implemented for the requested operation:

- do not report a destructive restore as completed
- do not report image capture as completed
- jobs requiring raw disk operations should remain waiting for the imaging node

This is a security boundary, not a cosmetic limitation.
