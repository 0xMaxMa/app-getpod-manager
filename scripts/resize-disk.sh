#!/bin/bash
set -e
DISK=$(lsblk -dn -o NAME,TYPE | awk '$2=="disk"{print "/dev/"$1; exit}')
[[ -n "$DISK" ]] || { echo "no disk found" >&2; exit 1; }
PART=$(findmnt -n -o SOURCE /)
PART_NUM=$(echo "$PART" | grep -o '[0-9]*$')
[[ -n "$PART_NUM" ]] || { echo "cannot detect root partition number from $PART" >&2; exit 1; }
sudo growpart "$DISK" "$PART_NUM"
sudo resize2fs "$PART"
