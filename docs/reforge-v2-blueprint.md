# ReForge v2 Product Blueprint

## 1. Product direction

ReForge v2 is an endpoint deployment, imaging, provisioning, and lifecycle platform for Windows, Linux, and macOS environments.

The design combines:

- FOG-style PXE registration, capture, restore, tasking, host groups, multicast, diagnostics, and deployment queues
- Microsoft WDS/MDT-style task sequences, unattended Windows setup, driver injection, application installation, Active Directory placement, and post-deployment automation
- Modern macOS provisioning, installer workflows, package/profile deployment, and Apple device enrollment
- A clean enterprise management UI with a dedicated privileged imaging-node boundary

The management plane remains unprivileged. Destructive disk operations are executed only by dedicated imaging nodes with explicit job authorization.

---

## 2. ReForge v2 navigation

### Dashboard
- Deployment pipeline
- Live deployment queue
- Recent activity
- System health
- Imaging-node health
- PXE service status
- Storage usage
- Image counts
- Failed-job alerts
- Quick deployment action

### Images
- Gold Images
- Clone Images
- ISO Library
- macOS Installers
- Image Versions
- Capture Jobs

### Deployment
- Hosts (PXE)
- Deployments
- Task Sequences
- Department Templates
- Host Groups
- Multicast Sessions
- Scheduled Tasks

### Management
- Applications
- Driver Packs
- Directory Services
- Updates
- Printers
- Scripts

### Monitoring
- Audit Logs
- Reports
- Activity
- Imaging Nodes
- Storage Nodes

### System
- Settings
- Admin Center

---

## 3. UI design system

ReForge v2 should use a restrained enterprise visual system.

### Layout
- Fixed left navigation
- Compact top command bar
- Wide content area
- Minimal nested cards
- Clear table-first management screens
- Responsive design for desktop/tablet
- Light, Dark, and System themes

### Top bar
- Global search
- Light / Dark / System selector
- Notifications
- Help
- Settings
- User/account menu

### Design rules
- No decorative gradients
- No excessive rounded cards
- No glow effects
- Use whitespace instead of boxes where possible
- Use tables for operational data
- Use cards only for summaries or grouped controls
- Keep destructive actions visually distinct
- Keep status colors semantic and limited

---

## 4. Login and identity

### Login screen
A simple sign-in page containing:

- ReForge logo
- Username
- Password
- Sign in
- Remember me
- Forgot password

### Future authentication
- Local accounts
- LDAP / Active Directory
- Microsoft Entra ID
- SAML/OIDC
- MFA
- Passkeys

### Admin Center
- Users
- Roles
- Permissions
- Authentication providers
- Password policy
- Session policy
- MFA requirements
- Branding
- Certificates
- Network
- Storage
- Imaging nodes
- PXE settings
- Audit retention

Recommended initial roles:

- Administrator
- Deployment Administrator
- Imaging Operator
- Help Desk
- Auditor
- Read Only

---

## 5. Imaging source types

ReForge v2 treats three image sources as first-class deployment objects.

### 5.1 ISO deployment

Purpose:
- clean OS installation
- Windows install media
- Linux install media
- recovery media
- vendor installers

ISO metadata:

- ID
- name
- OS family
- OS edition
- version
- architecture
- checksum
- size
- boot mode
- UEFI compatibility
- unattended configuration
- storage location
- created timestamp
- enabled state

Windows ISO workflow:

1. PXE boot into WinPE
2. Retrieve deployment authorization
3. Partition disk
4. Apply Windows image or installer source
5. Inject drivers
6. Apply unattend.xml
7. Reboot
8. Run post-install agent
9. Rename device
10. Join AD / Entra workflow
11. Install applications
12. Run updates
13. Validate
14. Complete

Linux ISO workflow:

1. PXE/iPXE
2. boot installer kernel/initrd
3. supply unattended configuration
4. partition/install
5. run post-install tasks
6. inventory device
7. mark deployment complete

---

### 5.2 Gold Images

Gold Images are generalized, maintained operating-system reference images.

Metadata:

- ID
- name
- OS
- version
- architecture
- image version
- image path
- format
- checksum
- size
- capture date
- source host
- Sysprep/generalization state
- supported hardware families
- notes
- lifecycle state

Lifecycle:

Draft -> Capturing -> Validating -> Active -> Superseded -> Archived

Gold Image workflow:

1. prepare reference host
2. generalize
3. PXE into imaging environment
4. authorize capture
5. detect source disk
6. capture
7. compress
8. checksum
9. upload to image storage
10. validate
11. publish version

---

### 5.3 Clone Images

Clone Images represent machine-oriented or hardware-family-oriented disk captures.

Use cases:

- labs
- kiosks
- appliance replacement
- identical workstation fleets
- rapid recovery

Metadata:

- source host
- source serial/model
- disk topology
- partition topology
- firmware mode
- encryption state
- filesystem types
- image checksum
- compatible hardware family
- capture timestamp

Restore validation should check:

- destination disk size
- UEFI/BIOS compatibility
- partition layout
- architecture
- hardware-family rules
- encryption requirements

---

## 6. Image storage architecture

ReForge separates image metadata from image data.

Supported storage targets should include:

- local filesystem
- NFS
- SMB
- S3-compatible object storage
- dedicated ReForge storage node

Image objects should support:

- checksums
- compression
- versioning
- replication
- retention policy
- verification scans
- capacity monitoring

Suggested formats by workflow:

- partclone-based filesystem images
- compressed raw block images where required
- WIM/ESD for Windows workflows
- installer ISO
- APFS-aware/macOS workflows where applicable

---

## 7. Task Sequence engine

Task Sequences are reusable ordered deployment workflows.

### Core step types

- Inventory host
- Validate firmware
- Validate disk
- Wipe disk
- Partition disk
- Format filesystem
- Apply ISO
- Restore image
- Inject drivers
- Set hostname
- Configure network
- Join Active Directory
- Move computer to OU
- Entra enrollment step
- Install application
- Install application bundle
- Install printer
- Run PowerShell
- Run shell script
- Run Windows Update
- Run Linux package updates
- Reboot
- Wait for agent
- Validate service
- Validate network
- Capture inventory
- Complete

Each step should support:

- enabled/disabled
- ordering
- timeout
- retries
- continue-on-error
- conditions
- platform filter
- architecture filter
- model filter
- department filter
- success/failure output

---

## 8. Deployment wizard

The primary deployment wizard should be:

1. Computer
2. Deployment Source
3. Task Sequence
4. Department/Profile
5. Applications
6. Drivers
7. Directory
8. Review
9. Queue

Deployment Source choices:

- ISO
- Gold Image
- Clone Image
- macOS Provisioning

Review screen should show the resolved plan before any destructive action.

---

## 9. PXE and boot services

### PXE capabilities

- iPXE HTTP boot
- BIOS
- UEFI x64
- ARM64 where supported
- existing-DHCP integration
- proxy DHCP option
- managed DHCP option
- unknown-host policy
- host registration
- one-time task boot
- normal local boot
- diagnostics
- memtest
- rescue environment

### Host boot menu

Known hosts:
- execute assigned deployment
- capture image
- diagnostics
- inventory
- local disk boot

Unknown hosts:
- quick registration
- full registration
- inventory only
- deny deployment

### Host identity

Use:

- MAC
- SMBIOS UUID
- serial number
- manufacturer
- model
- architecture
- TPM presence
- firmware type

---

## 10. Dedicated imaging node

The imaging node is the privileged data plane.

Responsibilities:

- PXE boot environment hosting
- disk discovery
- disk wipe
- partitioning
- image capture
- image restore
- compression/decompression
- checksum validation
- multicast
- storage transfer
- WinPE/Linux PE support
- driver staging

Security requirements:

- never expose arbitrary shell execution from the web API
- short-lived signed job authorization
- job ID
- target host identity
- allowed source image
- allowed target disk
- expiration
- nonce
- audit trail

States:

Online
Busy
Degraded
Offline
Draining

---

## 11. Hosts and inventory

Host record should expand beyond the current model.

Recommended fields:

- hostname
- MAC addresses
- serial number
- asset tag
- manufacturer
- model
- architecture
- CPU
- memory
- disks
- NICs
- firmware mode
- TPM
- OS
- OS version
- department
- host group
- location
- last seen
- enrollment state
- deployment state
- assigned task
- notes

### Host groups

Groups can be:

- static
- department-based
- model-based
- site-based
- dynamic query

Bulk actions:

- deploy
- assign profile
- schedule task
- install apps
- run inventory
- Wake-on-LAN
- move group

---

## 12. Applications

Application packages should support:

- MSI
- EXE
- MSIX
- PowerShell
- shell
- PKG for macOS
- custom package type

Fields:

- name
- version
- platform
- architecture
- installer
- install command
- uninstall command
- detection rule
- dependencies
- install order
- reboot behavior
- timeout
- required/optional
- checksum

### Bundles

Examples:

- Finance Standard
- Clinical Standard
- IT Engineering
- New Employee
- Design Workstation

---

## 13. Drivers

Driver Packs should be independent managed objects.

Fields:

- vendor
- model
- OS
- OS version
- architecture
- pack version
- source
- checksum
- status

Matching order:

1. exact manufacturer + model
2. hardware-family match
3. vendor fallback
4. generic OS drivers

Windows deployment should allow offline injection before first boot.

---

## 14. Directory Services

Current LDAPS/StartTLS configuration remains the foundation.

Expand to support:

- multiple directory profiles
- multiple domain controllers
- encrypted credentials
- connection health
- OU browser
- department OU mapping
- computer-name preview
- domain join policy
- delegated service accounts

Future:
- Microsoft Entra ID
- Hybrid Join
- Autopilot integration where appropriate

---

## 15. Updates

### Windows
- Windows Update stage
- WSUS support
- update rings
- reboot management
- update status reporting

### Linux
- apt
- dnf/yum
- distribution-specific update commands

### macOS
- softwareupdate integration
- MDM-driven OS update policy

---

## 16. FOG-inspired features

ReForge v2 should implement:

- quick host registration
- full host registration
- capture task
- deploy task
- inventory task
- wipe task
- diagnostics task
- host groups
- scheduled tasks
- multicast deployments
- Wake-on-LAN
- task history
- storage-node health
- image replication
- snapin-style application deployment
- host notes
- PXE menu customization

ReForge should implement these concepts with its own code, data model, terminology, and UI.

---

## 17. Microsoft deployment-inspired features

ReForge should implement:

- task sequences
- unattended Windows deployment
- driver injection
- application bundles
- OU placement
- computer naming rules
- post-install PowerShell
- state/status tracking
- deployment templates
- deployment validation
- reusable environment variables

ReForge remains its own platform and does not depend on deprecated WDS/MDT components.

---

## 18. macOS support

macOS should be presented as both provisioning and imaging because modern Apple hardware does not follow the same block-imaging model as traditional Windows PCs.

### Device classes

#### Intel Mac
Support where technically appropriate:

- network-assisted recovery workflows
- APFS-aware capture/restore workflows
- macOS installer workflows
- package deployment
- post-install scripts

#### Apple Silicon
Primary workflow:

- erase / restore
- install or reinstall macOS
- Apple Business Manager
- Automated Device Enrollment
- MDM enrollment
- configuration profiles
- application packages
- scripts/policies
- inventory

Traditional reusable block-clone workflows should not be assumed to be valid for Apple Silicon systems.

### macOS resources

- macOS Installer
- Apple Enrollment Profile
- Configuration Profile
- PKG Package
- Mac Application Bundle
- Mac Device
- Mac Provisioning Job

### Mac inventory

- serial
- hardware UUID
- model identifier
- Intel / Apple Silicon
- macOS version
- FileVault state
- enrollment state
- MDM state
- last seen

---

## 19. Admin Center

Sections:

### Users & Roles
- accounts
- roles
- permissions
- session revocation

### Authentication
- Local
- LDAP/AD
- Entra/OIDC
- MFA
- passkeys

### Infrastructure
- imaging nodes
- storage nodes
- PXE
- DHCP integration
- certificates
- DNS
- time/NTP

### Platform
- branding
- update channel
- retention
- audit
- notifications
- SMTP

### Security
- password policy
- MFA policy
- API tokens
- trusted networks
- secrets status
- audit retention

---

## 20. Reports

Initial reports:

- deployment success rate
- failed deployments
- deployment duration
- image usage
- image age
- host inventory
- hardware models
- OS versions
- application deployment status
- driver-pack usage
- PXE registrations
- imaging-node health
- storage utilization
- audit events

Export:

- CSV
- JSON
- PDF later

---

## 21. Database model expansion

Current models should remain and evolve.

### New recommended entities

- ImageSource
- ImageVersion
- ISOImage
- CloneImage
- CaptureJob
- TaskSequence
- TaskSequenceStep
- DeploymentTemplate
- DriverPack
- HostGroup
- HostGroupMember
- HostHardware
- ImagingNode
- StorageNode
- StorageLocation
- MulticastSession
- ScheduledTask
- DeploymentEvent
- ApplicationBundle
- Script
- Printer
- MacInstaller
- MacEnrollmentProfile
- ConfigurationProfile
- SecretReference
- Notification
- ReportDefinition

### DeploymentJob expansion

Add:

- source type
- source ID
- task-sequence ID
- imaging-node ID
- requested by
- scheduled time
- started time
- completed time
- result
- failure code
- failure detail
- current step ID
- reboot count
- destructive authorization state

---

## 22. API expansion

Recommended API groups:

- /api/images
- /api/images/gold
- /api/images/clones
- /api/images/isos
- /api/captures
- /api/task-sequences
- /api/deployments
- /api/hosts
- /api/host-groups
- /api/applications
- /api/application-bundles
- /api/drivers
- /api/directory
- /api/imaging-nodes
- /api/storage-nodes
- /api/reports
- /api/admin/users
- /api/admin/roles
- /api/admin/auth
- /api/admin/settings
- /api/macos
- /api/audit

Worker-only endpoints remain separated with dedicated authentication.

---

## 23. Dashboard specification

Dashboard layout:

### Header
- Dashboard title
- environment status
- ReForge node name/version
- Start Deployment
- View Logs

### Deployment Pipeline
Stages:

PXE Boot -> Select Source -> Drivers -> Directory -> Applications/Updates -> Complete

### Live Queue
Columns:

- Device
- Source
- Department
- Stage
- Progress
- Started
- Actions

### Recent Activity
Examples:

- Deployment started
- Image captured
- Image version activated
- Device registered
- Driver pack updated
- Department template modified
- Imaging node offline

### System Health strip
- API
- Database
- Worker
- PXE
- Imaging nodes
- Storage

---

## 24. Login screen specification

Visual structure:

- centered login panel
- ReForge logo/name
- "Deployment Platform"
- Username
- Password
- Remember me
- Sign in
- Forgot password

No dashboard data should be visible before authentication.

Optional organization branding:

- organization name
- logo
- support link

---

## 25. Phased implementation plan

### Phase 1 - UI and information architecture

1. new shell/navigation
2. login redesign
3. dashboard redesign
4. Light/Dark/System theme
5. Admin Center shell
6. ISO Library screen
7. Clone Images screen
8. Task Sequences screen
9. Drivers screen
10. Reports shell

### Phase 2 - data model and API

1. image-source abstraction
2. ISO model/API
3. Clone Image model/API
4. Image Version model/API
5. Task Sequence model/API
6. Driver Pack model/API
7. Host Group model/API
8. Imaging Node model/API
9. Storage Node model/API
10. deployment-event model

### Phase 3 - Windows deployment

1. WinPE environment
2. imaging-node agent
3. disk inventory
4. disk partition
5. ISO deployment
6. WIM/image deployment
7. driver injection
8. unattend generation
9. AD join
10. application stage
11. Windows Update
12. validation

### Phase 4 - capture/clone

1. capture authorization
2. disk capture
3. compression
4. checksum
5. upload
6. clone metadata
7. gold-image promotion
8. versioning
9. restore validation

### Phase 5 - FOG-style operations

1. quick registration
2. host groups
3. scheduled tasks
4. Wake-on-LAN
5. diagnostics
6. wipe task
7. multicast
8. storage replication

### Phase 6 - macOS

1. Mac inventory
2. macOS installer library
3. PKG catalog
4. profiles
5. Apple enrollment objects
6. MDM connector abstraction
7. Intel Mac supported capture/restore paths
8. Apple Silicon provisioning workflow

### Phase 7 - enterprise hardening

1. RBAC
2. external identity
3. MFA
4. secret vault abstraction
5. HA control plane
6. multiple imaging nodes
7. storage replication
8. notifications
9. reporting
10. backup/restore

---

## 26. v2 acceptance criteria

ReForge v2 reaches the first major milestone when an administrator can:

1. sign in through the new management UI
2. register a PXE host
3. upload/import an ISO
4. create or capture a Gold Image
5. create a Clone Image
6. create a Task Sequence
7. assign drivers/applications
8. queue a deployment
9. PXE boot a target
10. execute authorized imaging on a dedicated imaging node
11. track every step live
12. join a Windows endpoint to AD
13. install department applications
14. complete validation
15. review the full audit trail

A subsequent macOS milestone is reached when ReForge can inventory Macs, manage macOS installers/packages/profiles, and orchestrate modern Apple enrollment/provisioning workflows while supporting capture/restore only on hardware/workflows where that model is technically appropriate.

---

## 27. Architectural rule

The ReForge web/API control plane must never receive unrestricted raw-disk privileges.

The UI creates intent.
The API validates intent.
The worker orchestrates intent.
The imaging node performs only specifically authorized privileged operations.

That boundary remains a core ReForge design principle.
