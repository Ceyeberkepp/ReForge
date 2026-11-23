# Windows deployment identity and recovery safety gates

This document defines **acceptance criteria**, not currently implemented physical-disk functionality. The file-backed imaging-node tool does not perform Sysprep, PXE boot, AD join or physical restore. Do not expose a physical write path until signed, disk-specific authorization is implemented and tested.

## Two mutually exclusive workflows

| Property | New Windows endpoint deployment | Same-device disk recovery |
| --- | --- | --- |
| Source | Captured, Sysprep-generalized gold image | Recovery image for the **same** endpoint |
| Windows machine SID | Generated independently by Windows specialization on each target | Preserved as part of same-device restoration |
| Computer name | Unique, reserved for the target before deployment | Original name restored |
| AD computer account | Join/create or reset using an authorized domain workflow | Reconcile the existing account; do not create a duplicate |
| Intended targets | One or many distinct endpoints | Only the original approved endpoint |
| Failure behavior | Stop before disk write if image preparation or authorization is invalid | Stop if original endpoint/disk authorization is invalid |

**Machine SID** is not a Wi-Fi SSID or the AD computer-object SID. ReForge must not attempt to invent a machine SID or infer its uniqueness solely from hostname or AD account identity.

## Required new-endpoint gates

1. Record a trustworthy image-preparation attestation showing `Sysprep /generalize /oobe /shutdown` completed before image capture. A user-editable `sysprep_ready` flag alone is insufficient proof.
2. Reject captured Windows images not marked and verified as generalized. Do not clone an already-specialized running Windows installation to multiple endpoints.
3. Bind the approved job to deployment ID, endpoint identity, exact target disk stable identifier, operation, image digest, expiry and nonce. Verify a signed grant at the isolated imaging node immediately before destructive access; reject reuse and mismatches.
4. Require an operator to review the disk model, capacity, stable identifier and intended destructive operation. Fail closed on ambiguous or changing identifiers.
5. After first boot, Windows Setup must specialize the installation. Confirm unique computer names and AD computer accounts, and collect a trustworthy post-specialization machine SID for **distinct** Windows targets. A SID check must not rely on AD objectSID or SSID.
6. If identity verification is unavailable, report `unverified` and block a successful-completion claim. Never silently declare uniqueness.

## Required same-device recovery gates

Recovery must be an explicit mode with an image bound to the original device and a signed grant for its selected disk. Restoring the original machine SID is expected in this mode. Do not reuse this workflow to provision two simultaneously active Windows machines with the same identity.

## Test and evidence matrix

- **1 endpoint:** generalized Windows image specializes successfully; post-boot machine SID and hostname are recorded; AD join and required/optional software are verified.
- **2 endpoints:** same gold image, distinct machine SIDs, names and AD accounts; no duplicated endpoint credentials or tokens.
- **48 endpoints:** identity checks on all 48, unique target-disk grants, isolation between concurrent jobs, retries and cancellation. Measure per-endpoint elapsed time; the <=25-minute goal is **not verified** until representative physical hardware, network and image-size tests pass.
- **Negative cases:** non-generalized image; missing/expired/replayed signature; wrong disk; device mismatch; disk identifier changes; failed specialization; duplicate SID; failed AD join; interrupted restore. Each must fail safely with an audit record.
- **Recovery:** restore to the authorized original endpoint, verify the original identity is preserved, and reject another endpoint.

No acceptance result should be marked passed based solely on a mock GUI, static metadata, simulated PXE, or file-backed copy tests.
