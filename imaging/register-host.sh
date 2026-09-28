#!/usr/bin/env bash
set -Eeuo pipefail

API="${REFORGE_API_URL:?Set REFORGE_API_URL}"
IFACE="${REFORGE_IFACE:-$(ip route show default | awk '/default/ {print $5; exit}')}"
MAC="$(cat "/sys/class/net/${IFACE}/address")"
SERIAL="$(cat /sys/class/dmi/id/product_serial 2>/dev/null || true)"
MFG="$(cat /sys/class/dmi/id/sys_vendor 2>/dev/null || true)"
MODEL="$(cat /sys/class/dmi/id/product_name 2>/dev/null || true)"

curl -fsS -X POST "${API}/api/hosts/register"   -H 'Content-Type: application/json'   -d "$(python3 - <<PY
import json
print(json.dumps({
  "mac_address": "${MAC}",
  "serial_number": "${SERIAL}",
  "manufacturer": "${MFG}",
  "model": "${MODEL}"
}))
PY
)"
echo
