#!/bin/bash
for f in /sys/devices/system/memory/memory*/state; do
  [[ "$(cat "$f")" == "offline" ]] && echo online > "$f" || true
done
