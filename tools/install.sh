#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

curl -s -X POST http://localhost:10850/api/v1/apps/install \
  -H "X-Api-Key: $GATEWAY_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"local_path\": \"/home/dev/projects/app-getpod-manager\",
    \"env_vars\": {
      \"API_KEY\": \"${API_KEY}\",
      \"SSH_HOME\": \"/home/dev\"
    }
  }" | jq
