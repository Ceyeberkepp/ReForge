# KB-002 — ReForge web UI is unavailable

## Symptoms

- browser cannot reach `http://SERVER-IP:5173`
- connection reset
- HTTP 500 from Nginx
- local access works but remote access does not

## Check container status

```bash
cd /opt/reforge

docker compose --env-file .env -f infra/docker-compose.yml ps
```

Expected services include web, API, worker, and the selected SQL database.

## Check local web service

```bash
curl -I http://127.0.0.1:5173/
curl http://127.0.0.1:5173/health
ss -lntp | grep 5173
```

## Check logs

```bash
docker compose --env-file .env -f infra/docker-compose.yml logs --tail=100 web
docker compose --env-file .env -f infra/docker-compose.yml logs --tail=100 api
```

## If local works but remote fails

Check routing from the client and server. A successful TCP connect alone does not prove traffic is reaching the intended host.

On the ReForge host:

```bash
ip -br addr
ip route
```

From the client:

```powershell
Test-NetConnection SERVER-IP -Port 5173
tracert -d SERVER-IP
```

If packet capture on the ReForge interface sees no client packets, investigate upstream routing/NAT rather than changing ReForge.
