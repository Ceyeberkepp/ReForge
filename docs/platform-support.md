# Platform support

ReForge is hypervisor-independent because it runs inside a standard Linux guest or container rather than depending on a hypervisor API.

## Supported deployment targets

Recommended:

- Debian 12/13 VM
- Ubuntu Server 24.04 LTS VM
- Debian/Ubuntu LXC with nesting/container runtime support
- bare-metal Debian/Ubuntu
- Docker-capable Linux guest

Common hypervisors:

- Proxmox VE
- VMware ESXi / vSphere
- Microsoft Hyper-V
- Nutanix AHV
- KVM / libvirt
- XCP-ng
- VirtualBox for labs

## Container notes

A ReForge management container does not need raw block-device access.

The future imaging node is intentionally separate. It should run on a trusted deployment network with only the privileges required to serve boot artifacts and process explicitly authorized imaging jobs.

For LXC, enable the container-runtime features required by Docker/Podman. Do not give the web/API container direct host disk access.

## Network requirements

Management plane:

- TCP 5173 to the ReForge web UI
- outbound access for package/container updates as allowed by policy

PXE plane depends on the existing network design and can involve:

- DHCP / proxy-DHCP
- TFTP for the first-stage bootloader
- HTTP for iPXE menus and imaging artifacts

The PXE page in ReForge stores the environment-specific settings and generates the chain URL.
