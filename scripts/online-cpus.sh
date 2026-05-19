#!/bin/bash
for f in /sys/devices/system/cpu/cpu*/online; do echo 1 > "$f" 2>/dev/null || true; done
nproc
