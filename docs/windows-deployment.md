# Windows deployment design

A Windows deployment resolves to:

1. Gold image
2. Host hardware profile
3. Department profile
4. Required applications
5. Per-device optional applications
6. Computer naming rule
7. Active Directory OU
8. Printers and scripts
9. Windows Update
10. Validation

## Gold image preparation

A Windows gold image should be generalized with Sysprep before capture. ReForge stores a `sysprep_ready` flag as deployment metadata.

## Department profiles

Department profiles own the variable layer that should not be baked into every image:

- required software
- optional software catalog
- OU placement
- naming convention
- printers
- post-install scripts

This allows one maintained Windows 11 gold image to serve Finance, HR, IT, Clinical, and other departments without image sprawl.

## Active Directory

Directory settings store connection metadata only. Passwords supplied to the connection-test endpoint are not persisted. Production domain-join credentials should be kept in encrypted secret storage and exposed only to the privileged deployment worker.
