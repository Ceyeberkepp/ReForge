# KB-008 — Active Directory / LDAP troubleshooting

ReForge Directory Services supports secure LDAP connectivity.

Recommended:

- LDAPS
- LDAP StartTLS

Plain LDAP is disabled by default.

## Check network reachability

```bash
nc -vz DOMAIN-CONTROLLER 636
```

or for StartTLS/LDAP:

```bash
nc -vz DOMAIN-CONTROLLER 389
```

## ReForge fields

Verify:

- domain
- domain controller
- protocol
- port
- Base DN
- service account
- default OU

## Certificate issues

For LDAPS, the ReForge API container must trust the issuing CA. Do not work around production certificate failures by permanently disabling verification.

Use the built-in Directory Services connection test after correcting trust and credentials.
