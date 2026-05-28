#!/bin/sh
# df -P is POSIX — works on GNU coreutils and BusyBox (Alpine)
# Output is in 1K blocks; multiply by 1024 to get bytes for metrics.go conversion
df -P / | tail -1 | awk '{print $2*1024, $3*1024, $4*1024}'
