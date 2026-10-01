# KB-010 — Collect ReForge support data

Run from the ReForge server:

```bash
cd /opt/reforge

echo "===== VERSION ====="
git log -5 --oneline

echo "===== STACK ====="
docker compose --env-file .env -f infra/docker-compose.yml ps

echo "===== HEALTH ====="
curl -sS http://127.0.0.1:5173/health || true

echo "===== LISTENER ====="
ss -lntp | grep 5173 || true

echo "===== API LOG ====="
docker compose --env-file .env -f infra/docker-compose.yml logs --tail=100 api

echo "===== WORKER LOG ====="
docker compose --env-file .env -f infra/docker-compose.yml logs --tail=100 worker

echo "===== WEB LOG ====="
docker compose --env-file .env -f infra/docker-compose.yml logs --tail=100 web

echo "===== NETWORK ====="
ip -br addr
ip route

echo "===== STORAGE ====="
df -h
docker volume ls
```

For PXE cases also provide:

- BIOS or UEFI
- client VLAN/subnet
- DHCP/next-server configuration
- boot filename
- output of `curl http://127.0.0.1:5173/boot/ipxe`
- whether login prompt appears
- the account/group permissions involved

Do not share passwords, worker tokens, database passwords, session cookies, or the full contents of `.env`.
