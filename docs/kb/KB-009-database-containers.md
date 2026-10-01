# KB-009 — Database and container troubleshooting

## Stack status

```bash
cd /opt/reforge
docker compose --env-file .env -f infra/docker-compose.yml ps
```

## Supported SQL backends

- PostgreSQL
- MariaDB/MySQL
- Microsoft SQL Server

Check the selected profile:

```bash
grep -E '^(COMPOSE_PROFILES|REFORGE_DB_DRIVER)=' /opt/reforge/.env
```

Do not print secrets from `.env` into support tickets.

## API cannot start

```bash
docker compose --env-file .env -f infra/docker-compose.yml logs --tail=150 api
```

## Database container problems

Use the service name shown by `docker compose ps`, then inspect its logs.

PostgreSQL example:

```bash
docker compose --env-file .env -f infra/docker-compose.yml logs --tail=150 postgres
```

The API includes database startup retries, so a database that is still starting may delay API readiness briefly.
