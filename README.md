# ReForge

ReForge is a modern PXE imaging and endpoint deployment platform inspired by the flexibility of FOG Project and the workflow of Microsoft WDS/MDT.

## Goals

ReForge is designed around reusable deployment layers instead of maintaining a separate monolithic image for every department:

```
Gold Image
  -> Hardware/Driver Profile
  -> Department Profile
  -> Required Software
  -> Optional Software
  -> Active Directory / Entra configuration
  -> Post-deployment tasks
```

## Phase 1

The initial platform includes:

- Gold image catalog and version metadata
- Department deployment profiles
- Software/application catalog
- Host inventory
- Deployment jobs and status
- Active Directory connection settings
- Computer naming rules and OU assignment
- Modern web dashboard
- API-first architecture
- PostgreSQL persistence
- Docker Compose development stack

## Planned imaging capabilities

- BIOS and UEFI PXE/iPXE boot
- Host registration from PXE
- Image capture and restore
- Windows Sysprep-aware gold images
- Linux image support
- Driver packs
- Multicast deployment
- Wake-on-LAN
- Unattended Windows setup
- Post-image agent
- Domain join / OU placement
- Software installation and detection rules
- PowerShell and shell post-deployment scripts
- Deployment templates and visual workflow designer
- Audit logs and role-based access

## Repository layout

```
apps/
  api/        FastAPI control plane
  web/        React/Vite management UI
infra/
  docker-compose.yml
docs/
  architecture.md
```

## Development

Copy the example environment file and start the stack:

```bash
cp .env.example .env
docker compose -f infra/docker-compose.yml up --build
```

Default services:

- Web UI: http://localhost:5173
- API: http://localhost:8080
- API docs: http://localhost:8080/docs
- PostgreSQL: localhost:5432

## Security

ReForge must never store plaintext Active Directory passwords. The initial API accepts connection metadata only; encrypted secret storage and a dedicated privileged deployment worker are part of the next implementation phase.

## License

See LICENSE.
