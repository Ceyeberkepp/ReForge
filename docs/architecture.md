# ReForge Architecture

## Control plane

The ReForge control plane owns configuration and orchestration. It does not directly run privileged disk operations.

Core resources:

- Gold Images
- Departments
- Software Packages
- Hosts
- Deployment Jobs
- Directory Services
- Deployment Templates

## Deployment layers

A host deployment is resolved in this order:

1. Gold image
2. Hardware/driver profile
3. Department profile
4. Department-required software
5. Per-device optional software
6. Computer naming
7. Directory join / OU placement
8. Post-install tasks
9. Patch/update phase
10. Validation and completion

## Future privileged worker

A separate worker will handle:

- iPXE menu generation
- image capture and restore
- disk partitioning
- multicast
- Wake-on-LAN
- WinPE/Linux imaging environments
- SMB/NFS image storage
- Windows unattend generation
- domain join secrets
- PowerShell and shell task execution

This separation keeps the public API unprivileged and reduces the blast radius of the imaging engine.

## Directory integration

Initial target: on-prem Active Directory over LDAPS.

Stored configuration:

- domain FQDN
- domain controller(s)
- LDAPS port
- base DN
- service account username
- workstation OU
- optional department OU mappings

Credentials will be stored only in encrypted secret storage and passed to the deployment worker only when required.

Future targets:

- Microsoft Entra ID
- Hybrid AD + Entra
- LDAP-compatible directories
