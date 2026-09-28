# ReForge

ReForge is a modern PXE imaging and endpoint deployment platform inspired by Microsoft WDS/MDT workflows and the flexibility of FOG Project.

## What is implemented

ReForge now includes a working control-plane MVP:

- React management console
- FastAPI REST API
- PostgreSQL persistence
- Docker Compose deployment
- Gold image CRUD
- Department deployment profiles
- Required and optional software per department
- Starter software catalog
- Host inventory and MAC registration
- Deployment queue, progress and retry/cancel controls
- Deployment-plan resolver
- Active Directory / LDAP configuration
- LDAPS bind testing without storing the supplied password
- Department OU and naming rules
- iPXE menu generation
- PXE host lookup
- Separate deployment worker
- Linux hardware-registration helper
- Windows post-image bootstrap
- Installer script
- PXE/network documentation

## Deployment model

ReForge avoids maintaining a separate monolithic image for every department.

```
Gold Image
  -> Hardware / Driver Profile
  -> Department Profile
  -> Required Software
  -> Optional Software
  -> Computer Naming
  -> Active Directory / OU
  -> Printers
  -> Post-install Scripts
  -> Updates
  -> Validation
```

A single maintained Windows 11 gold image can therefore serve Finance, HR, IT, Clinical and other departments.

## Important imaging-node status

The management server does **not** currently execute destructive raw-disk operations.

Jobs move to **Waiting for imaging node** until a dedicated privileged imaging environment is attached. This is intentional: the web/API containers should never be able to wipe a disk simply because they can reach an endpoint.

The dedicated imaging-node adapter will own:

- target-disk discovery and confirmation
- partitioning
- image capture
- image restore
- compression
- checksums
- driver injection
- multicast
- imaging kernel/initramfs
- signed short-lived deployment tokens

## Install

On a Debian/Ubuntu server with Docker Engine and the Docker Compose plugin:

```bash
git clone https://github.com/Ceyeberkepp/ReForge.git
cd ReForge
sudo bash install.sh
```

Or start manually:

```bash
cp .env.example .env
docker compose -f infra/docker-compose.yml up -d --build
```

Default services:

- Web UI: http://SERVER:5173
- API: http://SERVER:8080
- API documentation: http://SERVER:8080/docs
- PostgreSQL: SERVER:5432

## PXE / iPXE

ReForge exposes its generated boot menu at:

```
http://SERVER:8080/boot/ipxe?api_url=http://SERVER:8080
```

Typical boot files:

- Legacy BIOS: `undionly.kpxe`
- UEFI x64: `ipxe.efi`

ReForge is designed to coexist with an existing DHCP server rather than requiring DHCP replacement. See `docs/pxe-setup.md`.

## Active Directory

Directory Services supports connection metadata for:

- domain FQDN
- domain controller
- LDAP or LDAPS
- port
- Base DN
- service account
- default computer OU

The UI can perform a real LDAP/LDAPS bind test. The test password is accepted only for that request and is not written to the database.

Department profiles can assign their own OU and computer naming rule such as:

```
FIN-{SERIAL}
HR-{SERIAL}
IT-{SERIAL}
```

## Software catalog

Applications are stored independently from gold images. ReForge can model packages such as:

- Adobe Acrobat Reader
- Microsoft 365
- Microsoft Teams
- Chrome
- Firefox
- 7-Zip
- VLC
- Power BI
- VS Code
- PuTTY
- WinSCP
- Zoom
- GlobalProtect
- NinjaOne

Each package can carry its installer location, version, install order, silent-install command, detection rule and uninstall command.

## Repository layout

```
apps/
  api/          FastAPI control plane
  web/          React/Vite management UI
  worker/       deployment orchestration worker

imaging/
  register-host.sh
  windows-postinstall.ps1

infra/
  docker-compose.yml
  dnsmasq.example.conf

docs/
  architecture.md
  pxe-setup.md
  windows-deployment.md
```

## Security model

ReForge must never persist plaintext Active Directory passwords. Production join credentials should be stored in encrypted secret storage and exposed only to the privileged deployment worker for the duration of an authorized deployment.

Raw disk operations should require a short-lived signed job token tied to the deployment, endpoint identity and selected target disk.

## Current version

Control-plane MVP: **0.2.0**

## License

See LICENSE.
