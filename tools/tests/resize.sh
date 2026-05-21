#!/bin/bash
[ -f "$(dirname "$0")/.env" ] && source "$(dirname "$0")/.env"

# Usage: DISK_GIB=50 CPU_CORES=4 MEMORY_MIB=4096 make resize
# At least one of DISK_GIB, CPU_CORES, MEMORY_MIB must be set.

body="{}"
parts=()
[ -n "$DISK_GIB" ]    && parts+=("\"disk_gib\": $DISK_GIB")
[ -n "$CPU_CORES" ]   && parts+=("\"cpu_cores\": $CPU_CORES")
[ -n "$MEMORY_MIB" ]  && parts+=("\"memory_mib\": $MEMORY_MIB")

if [ ${#parts[@]} -eq 0 ]; then
  echo "Error: set at least one of DISK_GIB, CPU_CORES, or MEMORY_MIB" >&2
  exit 1
fi

body="{$(IFS=,; echo "${parts[*]}")}"

curl -s -X POST \
  "${BASE_URL:-http://localhost:5990}/resize" \
  -H "Authorization: Bearer ${API_KEY}" \
  -H "Content-Type: application/json" \
  -d "$body" | jq
