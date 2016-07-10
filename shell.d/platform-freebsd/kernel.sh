#!/bin/sh
kernel_boot_epoch="$(sysctl -n kern.boottime | awk '{ print $4 }' | tr -d ',')"

echo "kernel.version:str:$(uname -K)"
echo "kernel.hostname:str:$(uname -n)"
echo "arch:str:$(uname -m)"
echo "uptime:int:$(( $(date +%s) - ${kernel_boot_epoch} ))"
echo "booted_at:date:$(date -j -f '%s' +'%Y-%m-%dT%H:%M:%S%z' ${kernel_boot_epoch})"
