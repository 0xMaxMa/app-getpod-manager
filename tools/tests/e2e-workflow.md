# e2e-workflow: Full Test Workflow for getpod-manager

End-to-end test covering setup → install → API tests → uninstall.

**Prerequisites:** claude-gateway running locally.

Set these shell variables once at the start — used throughout the steps below:

```bash
# Run from repo root
APP_NAME=$(grep '^name:' app.yaml | awk '{print $2}')
```

---

## 1. Environment Setup

### 1.1 App env — `<repo-root>/.env`

Used by the API container at install time. Create if not exists:

```bash
# Run from repo root
if [ ! -f .env ]; then
  printf "API_KEY (any secret string for gateway auth): "
  read -r api_key
  printf "USER_HOME (path to home dir with .ssh/authorized_keys, e.g. /home/dev): "
  read -r user_home
  cat > .env <<EOF
API_KEY=${api_key}
USER_HOME=${user_home}
EOF
  echo ".env created."
fi
```

**Where to find each value:**

| Variable | Where |
|---|---|
| `API_KEY` | Any secret string — used to authenticate API requests |
| `USER_HOME` | Home directory whose `.ssh/authorized_keys` the API will manage (e.g. `/home/dev`) |

### 1.2 Test env — `tools/tests/.env`

Used by the test scripts to connect to the API. Create if not exists:

```bash
# Run from repo root
if [ ! -f tools/tests/.env ]; then
  printf "API_KEY (same value as app .env): "
  read -r api_key
  cat > tools/tests/.env <<EOF
API_KEY=${api_key}
BASE_URL=http://localhost:10850/app/${APP_NAME}/api
EOF
  echo "tools/tests/.env created."
fi
```

> `BASE_URL` defaults to the gateway proxy path. Set it to `http://localhost:5990` to test the container port directly.

---

## 2. Uninstall (if exists)

```bash
gateway app uninstall ${APP_NAME} 2>/dev/null || true

# Verify containers are gone
docker ps --filter name=${APP_NAME}
# Expected: empty
```

---

## 3. Install

```bash
# Run from repo root
APP_DIR=$(pwd)
API_KEY=$(grep '^API_KEY=' .env | cut -d= -f2)
USER_HOME=$(grep '^USER_HOME=' .env | cut -d= -f2)

gateway app install --local "$APP_DIR" \
  --env "API_KEY=${API_KEY}" \
  --env "USER_HOME=${USER_HOME}"
```

**Expected logs:**
- `Symlinked <path> → ~/.claude-gateway/apps/${APP_NAME}`
- `Containers healthy`
- `Install complete: ...`

**Verify symlink:**

```bash
ls -la ~/.claude-gateway/apps/${APP_NAME}
# Expected: lrwxrwxrwx ... ${APP_NAME} -> /path/to/repo
```

**Verify containers:**

```bash
docker ps --filter name=${APP_NAME}
# Expected: ${APP_NAME}-api (healthy)
```

---

## 4. API Tests

All tests use scripts in `tools/tests/`. Run from that directory:

```bash
cd tools/tests
```

### 4.1 Health check (no auth)

```bash
make health
# Expected: {"status":"ok"} or similar
```

### 4.2 Metrics

```bash
make metrics
# Expected: JSON with cpu, memory usage data
```

### 4.3 SSH keys — list

```bash
make ssh-keys-list
# Expected: JSON array of current authorized keys
```

### 4.4 SSH keys — add

```bash
# Generate a throwaway test key (do not use a real key)
ssh-keygen -t ed25519 -f /tmp/test-e2e-key -N "" -q
SSH_KEY=$(cat /tmp/test-e2e-key.pub) make ssh-keys-add
# Expected: {"fingerprint":"SHA256:...","comment":"..."}

# Capture fingerprint for cleanup
TEST_FINGERPRINT=$(ssh-keygen -lf /tmp/test-e2e-key.pub | awk '{print $2}')
echo "Fingerprint: $TEST_FINGERPRINT"
```

### 4.5 SSH keys — verify added

```bash
make ssh-keys-list
# Expected: list includes the key added in 4.4
```

### 4.6 SSH keys — delete

```bash
FINGERPRINT="$TEST_FINGERPRINT" make ssh-keys-delete
# Expected: {"deleted":true} or 204 No Content
```

### 4.7 SSH keys — verify deleted

```bash
make ssh-keys-list
# Expected: list no longer contains the key from 4.4
```

### 4.8 Theme

```bash
THEME=dark make theme
# Expected: {"theme":"dark","colorTheme":"Default Dark Modern"}

THEME=light make theme
# Expected: {"theme":"light","colorTheme":"Default Light Modern"}
```

### 4.9 Resize (read-only check)

> ⚠️ This actually resizes VM resources. Run only when intentional.

```bash
# Example — resize disk to 30 GiB
DISK_GIB=30 make resize
# Expected: {"disk_gib":30,...} or success message

# Example — online additional CPUs
CPU_CORES=2 make resize

# Example — online additional memory
MEMORY_MIB=2048 make resize
```

---

## 5. Uninstall

```bash
gateway app uninstall ${APP_NAME}
```

**Verify symlink is removed:**

```bash
ls ~/.claude-gateway/apps/${APP_NAME}
# Expected: No such file or directory
```

**Verify containers are gone:**

```bash
docker ps --filter name=${APP_NAME}
# Expected: empty
```
