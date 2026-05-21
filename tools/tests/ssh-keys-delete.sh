#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

# Usage: FINGERPRINT="SHA256:..." make ssh-keys-delete
if [ -z "$FINGERPRINT" ]; then
  echo "Error: FINGERPRINT is required" >&2
  exit 1
fi

ENCODED=$(python3 -c "import urllib.parse, os; print(urllib.parse.quote(os.environ['FINGERPRINT']))" 2>/dev/null \
  || node -e "console.log(encodeURIComponent(process.env.FINGERPRINT))")

curl -s -X DELETE \
  "${BASE_URL:-http://localhost:5990}/ssh-keys/${ENCODED}" \
  -H "Authorization: Bearer ${API_KEY}" | jq
