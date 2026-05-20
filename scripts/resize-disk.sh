#!/bin/bash
set -e
DISK=$(lsblk -dn -o NAME,TYPE | awk '$2=="disk"{print "/dev/"$1; exit}')
[[ -n "$DISK" ]] || { echo "no disk found" >&2; exit 1; }
growpart "$DISK" 1
resize2fs "${DISK}1"
