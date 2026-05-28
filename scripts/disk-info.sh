#!/bin/bash
df -B1 --output=size,used,avail / | tail -1 | tr -s ' '
