# ReForge imaging node: first working increment

This standalone Go command performs **real byte-for-byte capture and restore against regular files**, with a SHA-256 manifest. It provides a safe integration/test foundation for the future privileged imaging node. It is **not** a PXE-ready, physical-disk imaging service.

## Local lab usage

Run from `imaging/node`:

```sh
go test ./...
go run . -action capture -source ./test-disk.bin -destination ./gold.raw
go run . -action verify -source ./gold.raw
go run . -action restore -source ./gold.raw -destination ./restored-disk.bin
cmp ./test-disk.bin ./restored-disk.bin
```

Create `test-disk.bin` with your own disposable sample content first. Each destination must be a **new** file; existing files are never overwritten. The raw image is accompanied by a `.json` manifest containing format, byte length, and SHA-256 checksum.

## Boundaries and next steps

- Block devices, symlink sources, special files, and existing destinations are rejected. This tool cannot wipe physical disks.
- The tool is not exposed as an HTTP endpoint and is not wired to the unprivileged orchestrator.
- Images are presently uncompressed; use file-backed lab tests only.
- For production: signed, short-lived job grants bound to deployment, host identity, target disk and operation; operator review of destructive actions; imaging node identity and mTLS; heartbeat and lease management; streamed progress; crash-consistent capture, image metadata persistence, and atomic storage publication; PXE boot environment.
- Restores should only reach a real block device **after** those authorization and safety gates exist, with explicit disk-selection and size checks.
