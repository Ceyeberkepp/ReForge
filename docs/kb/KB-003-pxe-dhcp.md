# KB-003 — PXE boot and DHCP troubleshooting

## ReForge PXE entry point

```text
http://SERVER:5173/boot/ipxe
```

## Typical boot files

- Legacy BIOS: `undionly.kpxe`
- UEFI x64: `ipxe.efi`

These values are configurable in **PXE / Network**.

## Existing DHCP

ReForge is designed to coexist with an existing DHCP server. Configure the appropriate next-server/TFTP/boot filename options on the network that serves PXE clients.

## Verify ReForge output

```bash
curl http://127.0.0.1:5173/boot/ipxe
```

## If the client never reaches ReForge

Check:

1. DHCP address assignment
2. VLAN scope
3. DHCP PXE options
4. TFTP/initial bootloader reachability
5. BIOS vs UEFI boot-file match
6. routing/firewall to ReForge
7. generated ReForge server URL

Do not replace enterprise DHCP unless that is an intentional design choice.
