#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

curl -s http://localhost:10850/api/v1/apps \
  -H "X-Api-Key: $GATEWAY_API_KEY" | jq
