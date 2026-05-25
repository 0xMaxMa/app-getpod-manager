#!/bin/bash
set -e
DISK=$(lsblk -dn -o NAME,TYPE | awk '$2=="disk"{print "/dev/"$1; exit}')
[[ -n "$DISK" ]] || { echo "no disk found" >&2; exit 1; }
PART=$(findmnt -n -o SOURCE /)
PART_NUM=$(echo "$PART" | grep -o '[0-9]*$')
sudo growpart "$DISK" "$PART_NUM"
sudo resize2fs "$PART"
