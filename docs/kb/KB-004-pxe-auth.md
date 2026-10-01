# KB-004 — PXE login and permission troubleshooting

ReForge requires credentials before deployment content is shown.

## Expected flow

```text
Boot menu
→ Install / Capture
→ Username/password
→ permission check
→ ISO / Gold / Clone / Capture choices
```

## If login fails

Verify the account is enabled and credentials are correct.

## If login works but no deployment source appears

Check group/user permissions. Relevant permissions include:

- `pxe.login`
- `pxe.install`
- `pxe.capture`
- `pxe.iso`
- `pxe.gold`
- `pxe.clone`
- `pxe.diagnostics`

A user who can sign in but lacks `pxe.iso`, for example, should not see ISO deployment options.

## Local Boot

Local Boot should remain separate from protected deployment content.
