#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

# Usage: THEME=dark make theme
#        THEME=light make theme
THEME="${THEME:-dark}"

curl -s -X PATCH \
  "${BASE_URL:-http://localhost:5990}/theme" \
  -H "X-Api-Key: ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d "{\"theme\": \"${THEME}\"}" | jq
