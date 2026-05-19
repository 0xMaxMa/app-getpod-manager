#!/bin/bash
df -BG --output=size,used,avail / | tail -1 | tr -s ' ' | sed 's/G//g'
