#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

curl -s -X GET \
  "${BASE_URL:-http://localhost:5990}/ssh-keys" \
  -H "Authorization: Bearer ${API_KEY}" | jq
