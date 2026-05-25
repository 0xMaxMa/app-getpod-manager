#!/bin/bash
for f in /sys/devices/system/cpu/cpu*/online; do echo 1 | sudo tee "$f" > /dev/null 2>&1 || true; done
nproc
