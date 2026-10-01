# KB-007 — Hosts, MAC addresses, serials, and hardware UUIDs

A MAC address is **optional** for manually managed ReForge hosts.

Supported identity fields include:

- hostname
- serial number
- SMBIOS/hardware UUID
- MAC address when available
- manufacturer
- model

## PXE clients

PXE firmware may provide:

- network-interface MAC
- SMBIOS UUID

ReForge can carry both through the PXE workflow.

## Why MAC is optional

MAC-only identity is fragile when:

- USB/network adapters are replaced
- docking stations change
- Wi-Fi and Ethernet interfaces differ
- hardware has multiple NICs

Serial number and hardware UUID provide additional identifiers.

## Minimum identity

A host registration should supply at least one usable identity such as hostname, serial number, hardware UUID, or MAC.
