# ReForge installer and upgrade behavior

## New installation

The one-click installer installs its dependencies, selects a primary database, generates credentials, and starts the Compose application. If the default web port `5173` is occupied, it automatically chooses the next available TCP host port. An explicit `REFORGE_WEB_PORT` can set a starting port (1024–65535).

## Re-running an existing installation

**Do not rotate application secrets, database passwords, or worker tokens during routine updates.** The installer now detects the existing `/opt/reforge/.env` (or `REFORGE_ROOT/.env`) and preserves that file and its database credentials. It does not print the original administrator password or generate a replacement. For existing pre-upgrade configurations it adds a `REFORGE_WEB_PORT=5173` default if missing.

The installer's current upgrade path executes `docker compose up -d --build`; it does not automatically create a verified backup or perform rollback. **Take a tested database/content backup before an upgrade** and verify service health and database integrity afterward. Avoid running the one-click installer on production without a maintenance window.

## Security and limits

The default UI is currently HTTP (not TLS), and the project is not yet production imaging-ready. Use a properly configured HTTPS reverse proxy and restrict access until TLS automation is implemented. The installer does not start a privileged imaging node or enable destructive disk operations. Host port probing is a best-effort check, and another process can take a selected port before Docker binds it.
