#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

curl -s "http://localhost:10850/api/v1/apps/jobs/$1" \
  -H "X-Api-Key: $GATEWAY_API_KEY" \
  | sed 's/\x1b\[[0-9;]*[a-zA-Z]//g; s/\x1b.//g' \
  | tr -d '\000-\010\013-\037'
