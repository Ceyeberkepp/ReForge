# ReForge

ReForge is a modern endpoint imaging, reimaging, PXE deployment and device-management platform. It combines the deployment workflow people remember from Microsoft WDS/MDT with FOG-style flexibility and a cleaner administration experience.

## Support and knowledge base

- **[ReForge Support Guide](docs/REFORGE-SUPPORT-GUIDE.md)** — installation, updates, architecture, PXE, content storage, RBAC, directory services, logs, and support triage.
- **[ReForge Knowledge Base](docs/kb/README.md)** — step-by-step articles for web access, PXE/DHCP, PXE authentication, content uploads, imaging workflows, host identity, directory services, database/container troubleshooting, and support-data collection.

## Architecture

- **Go 1.27 control plane** — REST API, authentication, audit events, PXE menu generation, configuration and orchestration
- **Go 1.27 worker** — isolated deployment queue worker
- **React 19 + TypeScript UI** — fast browser management console with hand-written CSS
- **Unprivileged Nginx** — same-origin web/API proxy and production static hosting
- **Selectable primary database** — PostgreSQL, MariaDB/MySQL or Microsoft SQL Server
- **Optional data services** — MongoDB and Neo4j profiles
- **Dedicated imaging-node boundary** — raw disk capture/restore is never executed by the web/API service

## What the UI manages

- Gold images and versions
- Managed devices and hardware inventory
- Department deployment profiles
- Required software by department
- Optional software during deployment
- Application installers and silent commands
- Active Directory / LDAP configuration
- LDAPS and StartTLS connection testing
- Computer naming policies
- Department OU placement
- Printers and post-install scripts
- Windows-style deployment task sequences
- PXE / iPXE settings
- Deployment queue and status
- Audit history
- Administrator password rotation

## Deployment flow

```
PXE / iPXE
   ↓
Device identity
   ↓
Gold image
   ↓
Drivers
   ↓
Department profile
   ↓
Required applications
   ↓
Optional applications
   ↓
Computer naming
   ↓
Active Directory / OU
   ↓
Printers / scripts
   ↓
Updates
   ↓
Validation
```

## One-command install

Supported target: a Docker-capable Debian/Ubuntu Linux VM, LXC/container, or bare-metal system.

```bash
curl -fsSL https://raw.githubusercontent.com/Ceyeberkepp/ReForge/main/install.sh | sudo bash
```

The installer:

1. bootstraps the ReForge repository
2. lets you choose PostgreSQL, MariaDB/MySQL, or SQL Server
3. optionally enables MongoDB and/or Neo4j
4. generates unique database, worker and administrator secrets
5. builds the Go API, Go worker and production web UI
6. starts ReForge
7. prints the web address and initial administrator password

Default web address:

```
http://SERVER-IP:5173
```

## Hypervisor support

ReForge is hypervisor-independent because it runs in the guest OS. It can be deployed on Linux guests hosted by:

- Proxmox VE
- VMware ESXi / vSphere
- Microsoft Hyper-V
- Nutanix AHV
- KVM / libvirt
- XCP-ng
- other hypervisors capable of running a supported Linux guest

See `docs/platform-support.md`.

## Databases

Primary transactional stores:

- PostgreSQL — recommended
- MariaDB / MySQL
- Microsoft SQL Server

Optional data services:

- MongoDB
- Neo4j
- future Redis/Valkey coordination/cache support

ReForge deliberately keeps authoritative deployment state in a transactional SQL store. Document and graph databases can run alongside it without weakening deployment consistency.

See `docs/database-support.md`.

## PXE

The PXE configuration is managed from **PXE / Network** in the web UI.

ReForge stores:

- enabled/disabled state
- ReForge server URL
- DHCP integration mode
- DHCP server
- next-server / TFTP address
- BIOS boot filename
- UEFI x64 boot filename
- boot menu timeout
- unknown-device policy

The generated iPXE entry point is:

```
http://SERVER:5173/boot/ipxe
```

ReForge can work with an existing DHCP server instead of forcing DHCP replacement.

## Security

Security-related design includes:

- authenticated administrator sessions
- Argon2id password hashing
- HTTP-only SameSite cookies
- administrator password rotation
- audit event history
- restricted CORS
- secure web response headers
- LDAPS and LDAP StartTLS
- plaintext LDAP disabled by default
- separate worker authentication token
- non-root backend/worker containers
- Linux capability dropping
- no direct database port exposure by default
- raw-disk operations isolated from the management service
- dependency and vulnerability scanning in GitHub Actions
- Dependabot update monitoring
- Trivy vulnerability, secret and misconfiguration scanning

See `docs/security-compliance.md`.

## ReForge v2 blueprint

The planned ReForge v2 product direction, UI structure, ISO/Gold/Clone workflows, task sequences, imaging-node architecture, FOG/MDT-inspired capabilities, and macOS provisioning strategy are documented in `docs/reforge-v2-blueprint.md`.

## Compliance

ReForge is designed to **support** environments working toward SOC 2, HIPAA and GDPR requirements. No software repository by itself makes an organization compliant; operational controls, risk assessments, hosting, identity configuration, policies, evidence, contracts, retention and independent assessments still matter.

## Source layout

```
apps/
  api-go/       Go control plane
  worker-go/    Go deployment worker
  web/          React + TypeScript management UI

imaging/
  register-host.sh
  windows-postinstall.ps1

infra/
  docker-compose.yml
  dnsmasq.example.conf

docs/
  architecture.md
  database-support.md
  platform-support.md
  pxe-setup.md
  security-compliance.md
  windows-deployment.md
```

## Current imaging-node boundary

The management platform is working as the control plane, but it intentionally does not grant its web/API containers raw disk access.

Jobs that require physical image capture or restore remain in **Waiting for imaging node** until the dedicated imaging environment is attached. That imaging node will require explicit authorization for the deployment ID, endpoint identity and target disk rather than accepting arbitrary disk-write commands from the management network.

## License

See LICENSE.
