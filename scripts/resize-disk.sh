#!/bin/bash
set -e
DISK="${1:-}"
[[ "$DISK" =~ ^/dev/(sd|vd)[a-z]$ ]] || { echo "invalid device: $DISK" >&2; exit 1; }
growpart "$DISK" 1
resize2fs "${DISK}1"
