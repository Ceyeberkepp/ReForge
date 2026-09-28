# ReForge Imaging Node

This directory contains the endpoint-side and imaging-environment pieces.

## Current

- hardware registration script
- iPXE control-plane endpoints
- deployment-plan endpoint
- Windows post-image bootstrap
- queued worker orchestration
- safe mode that does not write raw disks

## Imaging adapter boundary

A production imaging node will run separately from the web/API containers and receive an explicit deployment job. It will be responsible for:

- enumerating the target disk
- wiping/partitioning only the selected target
- capture and restore
- compression and checksums
- image storage
- driver injection
- multicast
- boot artifacts

Raw disk actions must never be executed merely because the web API is reachable. The node should require a short-lived signed job token tied to the host MAC, deployment ID and target disk.
