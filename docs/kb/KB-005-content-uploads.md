# KB-005 — Upload Gold Images, ISOs, drivers, and applications

ReForge stores deployment files in persistent content storage rather than in SQL.

Default container path:

```text
/var/lib/reforge/content
```

Persistent Docker volume:

```text
reforge-content
```

## Supported types

### Gold Images
`.wim`, `.esd`, `.img`, `.gz`, `.zst`, `.zip`

### ISOs
`.iso`

### Driver Packs
`.zip`, `.cab`, `.7z`

### Applications
`.msi`, `.exe`, `.msix`, `.appx`, `.pkg`, `.zip`, `.ps1`, `.sh`

## Upload behavior

The browser shows upload progress. Nginx streams large requests to the API instead of buffering the whole file in the web container.

## Check storage

```bash
docker volume ls | grep reforge-content
df -h
```

## Troubleshooting

If an upload fails:

```bash
cd /opt/reforge
docker compose --env-file .env -f infra/docker-compose.yml logs --tail=100 web
docker compose --env-file .env -f infra/docker-compose.yml logs --tail=100 api
```

Confirm sufficient free disk space before retrying a large ISO/image.
