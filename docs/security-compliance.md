# ReForge security and compliance design

ReForge is engineered to support regulated deployments, but application code alone cannot make an organization SOC 2, HIPAA, or GDPR compliant. Compliance also depends on hosting, configuration, policies, workforce procedures, contracts, risk analysis, evidence, retention, incident response, and independent assessment where applicable.

## Built-in control direction

- authenticated administrative API
- Argon2id password hashing
- HTTP-only SameSite session cookies
- role-ready user model
- append-only application audit events
- secure response headers
- restricted CORS
- no plaintext directory password persistence
- non-root distroless backend container
- dedicated imaging-node security boundary
- raw-disk operations excluded from the management server
- database health checks and retry logic
- short-lived sessions
- dependency version pinning
- signed deployment-token design
- encrypted secret-provider integration point
- retention and deletion design
- centralized logging/SIEM integration point

## HIPAA-aligned controls

The design supports access control, authentication, audit controls, integrity safeguards, transmission security, backup/recovery, and least privilege. A covered entity or business associate must still perform and document its own risk analysis and safeguards.

## GDPR-aligned controls

The design supports data minimization, auditable access, encryption, retention controls, resilience, restoration, least privilege, and security testing. Deployments still need lawful-basis, controller/processor, retention, data-subject, and organizational procedures.

## SOC 2-aligned controls

The design supports evidence around security, availability, processing integrity, confidentiality, and privacy through access management, audit history, change control, service health, vulnerability scanning, backups, secure configuration, and operational evidence.

## Production checklist

1. enable HTTPS with managed certificates
2. replace bootstrap credentials
3. configure MFA/SSO
4. connect a supported secret manager
5. enable encrypted backups and restore tests
6. forward logs to a SIEM
7. isolate PXE/imaging networks
8. set retention/deletion policies
9. run SAST, dependency, secret and image scans
10. perform penetration testing
11. document risk analysis and operating procedures
