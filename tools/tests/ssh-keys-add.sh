#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

# Usage: SSH_KEY="ssh-ed25519 AAAA... comment" make ssh-keys-add
if [ -z "$SSH_KEY" ]; then
  echo "Error: SSH_KEY is required" >&2
  exit 1
fi

curl -s -X POST \
  "${BASE_URL:-http://localhost:5990}/ssh-keys" \
  -H "X-Api-Key: ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"key\": $(echo "$SSH_KEY" | jq -Rs .)}" | jq
