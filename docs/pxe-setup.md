# PXE / iPXE setup

ReForge generates its iPXE entry point at:

`GET /boot/ipxe`

Example chain target:

```
http://REFORGE-SERVER:5173/boot/ipxe?api_url=http://REFORGE-SERVER:5173
```

The example uses the default web port, **5173**. On installations where the one-click installer selected another available port, use the saved `REFORGE_WEB_PORT` from the installation `.env` instead. The internal API container's port 8080 is not published directly by the default Docker Compose deployment. This endpoint generates a chain script; it does not by itself provide a bootable imaging kernel or disk-write authorization.

## Existing DHCP server

You do not need ReForge to replace an existing DHCP server. Configure your current DHCP/PXE environment to boot iPXE and chain the ReForge URL.

Typical firmware paths:

- Legacy BIOS: undionly.kpxe
- UEFI x64: ipxe.efi

ReForge should coexist with the production DHCP server. The exact DHCP options depend on the existing platform and network design.

## Imaging environment

The current project includes the control-plane and worker orchestration path. The next privileged imaging-node adapter will supply:

- Linux imaging kernel and initramfs
- disk discovery and partitioning
- image capture/restore
- SMB/NFS or object-backed image storage
- multicast transport
- Windows post-image bootstrap
- driver injection
- unattended setup generation

The worker defaults to safe orchestration mode and does not run disk-destructive commands.
