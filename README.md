# WoL

A small, self-hosted Wake-on-LAN web app. One static binary, no external dependencies, one JSON file for data. Works well on a phone.

*Polski: [README.pl.md](README.pl.md)*

> Status: early version (0.x). The name is a working title.

## Why this exists

Waking a computer that has been off for a long time is less reliable than it looks, and the three usual ways of sending the magic packet behave differently from network to network:

- **Unicast to the computer's IP** only arrives while the router still has an ARP entry for the sleeping computer. After a long time off the entry is gone and the packet has nowhere to go.
- **Directed broadcast** (for example `192.168.1.255`) depends on a correct netmask and on the router being willing to forward it.
- **Limited broadcast** (`255.255.255.255`) is what the plain `wakeonlan` command sends. In many networks it works when the other two do not.

An app that sends only one of these, once, fails silently in some setups, and the fix is rarely obvious. This app sends **all of them in every attempt**, **retries** until the computer answers or a timeout is reached, and **shows exactly what was sent**, so a failure can be diagnosed instead of guessed at. Each device can choose its own methods.

## Features

- Wake by MAC address with up to three methods at once: limited broadcast on every network interface, directed broadcast, unicast.
- Retries every few seconds until the computer is online (ping and/or TCP ports) or the timeout is reached.
- Live progress and a per-attempt list of every packet target and result.
- Online status for each device, refreshed in the background.
- Per-device settings: methods, network interface, netmask, check ports, timeout, retry interval.
- Responsive page for phones and desktops, light and dark theme, English and Polish.
- Account with a hashed password, login rate limiting, strict security headers.
- Single static binary of about 7 MB. Standard library only, so there is nothing to keep updated except Go itself.

## Quick start (Docker Compose)

```yaml
services:
  wol:
    build: .
    container_name: wol
    restart: unless-stopped
    network_mode: host
    read_only: true
    cap_drop: [ALL]
    cap_add: [CHOWN, SETUID, SETGID]
    security_opt: ["no-new-privileges:true"]
    environment:
      WOL_ADDR: ":8080"
    volumes:
      - wol-data:/data

volumes:
  wol-data:
```

```sh
docker compose up -d --build
docker compose logs wol
```

`network_mode: host` is required. In Docker's default bridge network the broadcast packets never reach your LAN.

On the first start no account exists. The log prints a one-time **setup token**:

```
FIRST RUN: no account exists yet. Open the web page and enter this setup token: ...
```

Open `http://192.168.1.10:8080` (the address of the machine running the container), enter the token and choose a username and a password of at least 10 characters. There is no default password. The token is kept in memory only and is gone once the account exists.

Without Docker: `go build -o wol . && ./wol` (Go 1.24 or newer).

## Adding a device

1. **Name** and **MAC address** are required.
2. **IP address or host name** is optional but recommended. It is used for unicast packets and to check whether the computer is up. A host name is resolved with DNS and the last known address is remembered, so waking still works while the computer is off and the name no longer resolves.
3. Under **Advanced** you can pick the methods, the network interface, the netmask for an extra directed broadcast, ping and TCP ports for the online check, the timeout and the retry interval.

Without a host and a check the app cannot tell whether the computer woke up. It then sends three attempts and reports "sent".

## Configuration

| Variable | Default | Meaning |
|---|---|---|
| `WOL_ADDR` | `:8080` | Listen address, for example `127.0.0.1:8080` |
| `WOL_DATA` | `./data` (`/data` in the image) | Directory with `state.json` |
| `WOL_COOKIE_SECURE` | `false` | Set to `true` when served over HTTPS |
| `WOL_TRUST_PROXY` | `false` | Trust `X-Forwarded-For` and `X-Forwarded-Host` from a reverse proxy |
| `WOL_PORTS` | `9,7` | UDP ports the magic packet is sent to |
| `WOL_STATUS_INTERVAL` | `15s` | How often online status is refreshed (minimum `5s`) |
| `WOL_UID` / `WOL_GID` | `65532` | User and group the app switches to after startup (must not be 0) |

## Security

- Do not expose the app directly to the internet. Put it behind an HTTPS reverse proxy or reach it over a VPN.
- Behind a proxy set `WOL_COOKIE_SECURE=true` and `WOL_TRUST_PROXY=true`, and bind to the loopback address (`WOL_ADDR=127.0.0.1:8080`). Enable `WOL_TRUST_PROXY` only when exactly one trusted proxy sits in front, otherwise the client address can be spoofed and rate limiting loses its meaning.
- Example with Caddy:

  ```
  wol.example.com {
      reverse_proxy 127.0.0.1:8080
  }
  ```
- The container starts as root for a moment only to hand `/data` to an unprivileged user (`65532`), then switches to that user before opening any port. It keeps no privileges and cannot regain root. For this the container needs the `CHOWN`, `SETUID` and `SETGID` capabilities (already set in the Compose file above), and nothing else.
- Passwords are stored as PBKDF2-HMAC-SHA256 (600,000 iterations) with a random salt. Session tokens are random, stored only as hashes and expire after 30 days. The cookie is `HttpOnly` and `SameSite=Strict`.
- After 5 failed logins an address is blocked for 15 minutes.
- Requests that change anything must be JSON from the same origin. A strict Content Security Policy forbids inline scripts.
- The app never runs external commands and stores no credentials of the computers it wakes.
- `state.json` has mode `0600` and holds password hashes, so treat it (and its backups) as a secret.

**Lost password:** stop the app, set `"users": []` and `"sessions": []` in `state.json`, start it again and use the new setup token from the log. Devices are kept.

## Online check

The app pings the computer (ICMP) and optionally connects to TCP ports you list. A TCP port that answers with "connection refused" still counts as online, because only a running machine can refuse.

Ping works without privileges on Linux when the kernel allows it for the user the app runs as:

```sh
sysctl net.ipv4.ping_group_range
# if it does not include the container user (65532):
sudo sysctl -w net.ipv4.ping_group_range="0 2147483647"
```

If ping is not available the log says so and TCP ports are used instead.

## If a computer does not wake up

1. Enable Wake-on-LAN in the BIOS/UEFI (names vary: "Power On By PCI-E", "Resume by LAN") and disable "ErP" or "Deep Sleep" modes.
2. In the operating system allow the network card to wake the computer. On Windows that is the card's *Power Management* tab, and "Fast startup" should be off.
3. Prefer a wired connection. Waking over Wi-Fi works only on some cards.
4. Open **Details** on the device card after pressing *Wake*. It lists every target that was tried. Try the methods one by one in the advanced settings to find the one that works in your network.
5. Make sure the app runs in the same network as the computer. Broadcast packets do not cross routers unless the router is configured to forward them.
6. If the log says `permission denied` for `/data`, or `setuid`/`setgid` fails, the container is missing those three capabilities (`cap_add: [CHOWN, SETUID, SETGID]` together with `cap_drop: [ALL]`).

## Limitations

- IPv4 only.
- No mDNS (`.local`) name resolution. DNS names and IP addresses work.
- One account, created at first start. The password can be changed in the app.
- No built-in HTTPS, use a reverse proxy.

## Command line

```sh
wol                          # start the web app
wol send AA:BB:CC:DD:EE:FF   # send one magic packet to 255.255.255.255:9
wol send -addr 192.168.1.255:9 AA:BB:CC:DD:EE:FF
wol healthcheck              # exit code 0 when the web app answers
wol version
```

## Development

```sh
go vet ./...
go test -race ./...
WOL_DATA=./data go run .
```

The code is split by responsibility: `wake.go` (methods, retries), `probe.go` and `icmp*.go` (online check), `store.go` (JSON file), `auth.go` (passwords, rate limiting), `server.go` (HTTP API and security headers), `web/` (the page, embedded into the binary).

## License

Not chosen yet.
