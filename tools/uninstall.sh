#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

curl -s -X DELETE http://localhost:10850/api/v1/apps/getpod-manager \
  -H "X-Api-Key: $GATEWAY_API_KEY" | jq
