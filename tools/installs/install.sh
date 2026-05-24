#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

APP_DIR="$(cd "$(dirname "$0")/../.." && pwd)"
APP_NAME=$(grep '^name:' "${APP_DIR}/app.yaml" | awk '{print $2}')

body=$(jq -n \
  --arg path "$APP_DIR" \
  --arg key  "${API_KEY}" \
  --arg ssh  "${SSH_HOME:-/home/ubuntu}" \
  '{"local_path":$path,"env_vars":{"API_KEY":$key,"SSH_HOME":$ssh}}')

curl -s -X POST http://localhost:10850/api/v1/apps/install \
  -H "X-Api-Key: $GATEWAY_API_KEY" \
  -H "Content-Type: application/json" \
  -d "$body" | jq
