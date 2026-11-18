# ReForge imaging workflows: reimaging vs same-device recovery

ReForge must expose **two distinct operations** with different identity policies.

## 1. Reimage / Deploy (new Windows installation)

Use a **generalized Gold Image** (or an unattended installation source), not a raw clone from another live endpoint.

1. Prepare the Windows reference installation and approved applications.
2. Run the supported Windows Sysprep `/generalize /oobe /shutdown` workflow before capturing it. Validate the image has actually been generalized; the current `sysprep_ready` API field is an operator-provided declaration, not automated proof.
3. Capture the generalized image, maintain a manifest, and publish it as a Gold Image.
4. Enroll a target device, authorize imaging to the explicitly selected disk, and deploy.
5. Windows Setup/specialize establishes a distinct machine SID and computer identity on first boot.
6. Apply a unique hostname, AD domain join/OU placement, device drivers, software, and endpoint management enrollment.
7. Verify the endpoint identity and domain join before marking a deployment successful.

**Do not copy a live Windows installation to multiple computers without generalizing it.** Windows SID is not a Wi-Fi SSID. A unique Wi-Fi network SSID is not required per device; the requirement here is each Windows installation's **machine SID** and device identity. The current code prevents newly queued ungeneralized Windows Gold Images, but the capture-to-Sysprep-to-PXE workflow is not yet implemented end-to-end.

## 2. Clone / Recover (failed SSD/HDD on the *same* device)

Clone backups are distinct from Gold Images. Purpose: recover the same enrolled machine after a storage failure, retaining its original Windows installation and identity.

1. Capture a point-in-time backup of the device's storage, with partition layout, disk size, firmware mode, host ID, serial/model and encryption/recovery-key policy.
2. Store a checksum-verified restore image with backup metadata and device-binding information.
3. Replace the failed disk. Select the same managed host/device ID and an authorized replacement disk.
4. Restore only with explicit operator authorization, destination size checks, and a warning that the entire target disk will be overwritten.
5. Verify boot, filesystem, BitLocker recovery requirements, and endpoint identity.

**Do not distribute the same un-generalized clone to other active machines.** Doing so duplicates machine identity and can break domain and management enrollment. TPM/BitLocker binding may require recovery keys and re-enablement after restoring to replacement storage.

## Current implementation boundary

The initial `imaging/node` implementation is file-backed only and never writes block devices. Neither workflow is yet a physical PXE-to-disk solution. Pending: signed device- and disk-scoped grants, approved destructive actions, hardware inventory, captured partition tables, consistency checks, verified image catalog, WinPE/Linux imaging boot, and real node orchestration.
