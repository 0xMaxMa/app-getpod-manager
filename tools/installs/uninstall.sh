#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

APP_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
APP_NAME=$(grep '^name:' "${APP_DIR}/app.yaml" | awk '{print $2}')

curl -s -X DELETE "http://localhost:10850/api/v1/apps/${APP_NAME}" \
  -H "X-Api-Key: $GATEWAY_API_KEY" | jq
