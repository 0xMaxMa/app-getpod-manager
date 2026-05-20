# getpod-manager

A lightweight sidecar app that runs inside a GetPod VM. It exposes a JSON API for the GetPod control plane to manage resources on the VM — without needing direct SSH access.

## What it does

| Feature | Endpoint |
|---|---|
| Health check | `GET /health` |
| VM metrics (CPU / RAM / Disk) | `GET /metrics` |
| Resize disk / CPU / memory | `POST /resize` |
| List SSH keys | `GET /ssh-keys` |
| Add SSH key | `POST /ssh-keys` |
| Delete SSH key | `DELETE /ssh-keys/:fingerprint` |

All endpoints except `/health` require `Authorization: Bearer <API_KEY>`.

`API_KEY` is a secret you set at install time — it's used to authenticate requests to the getpod-manager API itself.

## How it fits in GetPod

```
GetPod Control Plane
      │
      │  HTTP (via reverse proxy)
      ▼
 getpod-manager  (running inside the VM as a gateway app)
      │
      ├── reads /proc/stat, /proc/meminfo  →  metrics
      ├── writes ~/.ssh/authorized_keys    →  ssh key management
      └── calls gateway scripts            →  resize disk / CPU / memory
```

## Install

Install via the claude-gateway API — either from the registry or directly from GitHub:

**From registry:**

```bash
curl -s -X POST http://localhost:10850/api/v1/apps/install \
  -H "X-Api-Key: <claude-gateway-admin-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "registry_app": "getpod-manager",
    "env_vars": {
      "API_KEY": "<secret>",
      "SSH_HOME": "/home/ubuntu"
    }
  }'
```

**From GitHub:**

```bash
curl -s -X POST http://localhost:10850/api/v1/apps/install \
  -H "X-Api-Key: <claude-gateway-admin-api-key>" \
  -H "Content-Type: application/json" \
  -d '{
    "github_url": "https://github.com/0xMaxMa/app-getpod-manager",
    "commit": "<commit-hash>",
    "env_vars": {
      "API_KEY": "<secret>",
      "SSH_HOME": "/home/ubuntu"
    }
  }'
```

| Env var | Required | Description |
|---|---|---|
| `API_KEY` | yes | Secret key to authenticate requests to the getpod-manager API |
| `SSH_HOME` | yes | Home directory of the user whose `authorized_keys` to manage (e.g. `/home/ubuntu`) |

Poll the install job:

```bash
curl -s http://localhost:10850/api/v1/apps/jobs/<jobId> \
  -H "X-Api-Key: <claude-gateway-admin-api-key>" | jq
```

## Uninstall

```bash
curl -s -X DELETE http://localhost:10850/api/v1/apps/getpod-manager \
  -H "X-Api-Key: <claude-gateway-admin-api-key>"
```

## API Reference

### GET /metrics

```json
{
  "cpu":    { "cores": 4, "usage_percent": 12.3 },
  "memory": { "total_mb": 8192, "used_mb": 2048, "free_mb": 6144 },
  "disk":   { "total_gb": 80, "used_gb": 20, "free_gb": 60 }
}
```

### POST /resize

Bring live-added resources online (after the hypervisor has already allocated more).

```json
{ "disk_gib": 80, "cpu_cores": 4, "memory_mib": 8192 }
```

All fields are optional — send only what you want to apply.

### GET /ssh-keys

```json
[
  { "fingerprint": "SHA256:...", "comment": "user@host", "raw": "ssh-ed25519 ..." }
]
```

### POST /ssh-keys

```json
{ "key": "ssh-ed25519 AAAA... user@host" }
```

### DELETE /ssh-keys/:fingerprint

URL-encode the fingerprint (e.g. `SHA256%3A...`).

---

## Local Development

Requirements: Go 1.22+, Docker, `jq`, access to a running claude-gateway instance.

### Setup

```bash
# 1. Copy env file
cp .env.example .env
# Edit .env — set API_KEY, SSH_HOME

# 2. Set gateway credentials for tools
echo "GATEWAY_API_KEY=<your-gateway-key>" > tools/.env
```

### Run locally with Docker Compose

```bash
cd tools

make start   # docker compose up -d
make logs    # follow logs
make stop    # docker compose down
```

### Deploy to gateway for testing

```bash
cd tools

make install-watch   # install + poll until completed
make stat            # check app status
```

### Poll a job manually

```bash
cd tools

make job id=<jobId>
```

### Available make targets

```
make help
```
