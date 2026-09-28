# ReForge database strategy

## Primary transactional database

ReForge supports these authoritative SQL stores:

- PostgreSQL
- Microsoft SQL Server
- MySQL
- MariaDB

PostgreSQL is the default. The selected primary database stores devices, images, department profiles, deployment state, users, sessions, PXE configuration, and audit records.

## Optional data services

The setup wizard also supports optional services for specialized workloads:

- MongoDB — document/event extensions
- Neo4j — network, device, and dependency relationship analysis
- Redis/Valkey — caching and distributed coordination

ReForge intentionally keeps authoritative deployment state in a transactional SQL database. A graph or document store can be enabled alongside it without changing the consistency rules for imaging jobs.

## Environment

Set:

`REFORGE_DB_DRIVER=postgres|mysql|mariadb|sqlserver`

and:

`REFORGE_DB_DSN=...`

The installer and web setup experience generate these settings for the selected database.
