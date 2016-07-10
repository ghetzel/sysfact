#!/bin/sh

swap_total_bytes="$(sysctl -n vm.swap_total)"
swap_used_bytes="$(( $(pstat -s | tail -n +2 | awk '{ print $3 }') * 1024 ))"
mem_pagesize="$(sysctl -n hw.pagesize)"

vmem_total_bytes="$(( $(sysctl -n vm.stats.vm.v_page_count) * ${mem_pagesize} ))"
vmem_wired_bytes="$(( $(sysctl -n vm.stats.vm.v_wire_count) * ${mem_pagesize} ))"
vmem_active_bytes="$(( $(sysctl -n vm.stats.vm.v_active_count) * ${mem_pagesize} ))"
vmem_inactive_bytes="$(( $(sysctl -n vm.stats.vm.v_inactive_count) * ${mem_pagesize} ))"
vmem_cached_bytes="$(( $(sysctl -n vm.stats.vm.v_cache_count) * ${mem_pagesize} ))"
vmem_freepage_bytes="$(( $(sysctl -n vm.stats.vm.v_free_count) * ${mem_pagesize} ))"

mem_total="$(sysctl -n hw.physmem)"
mem_free="$(( ${vmem_inactive_bytes} + ${vmem_cached_bytes} + ${vmem_freepage_bytes} ))"
mem_used="$(( ${mem_total} - ${mem_free} ))"

echo "memory.pagesize:int:${mem_pagesize}"
echo "memory.total:int:${mem_total}"
echo "memory.used:int:${mem_used}"
echo "memory.free:int:${mem_free}"
echo "memory.swap.total:int:${swap_total_bytes}"
echo "memory.swap.used:int:${swap_used_bytes}"
echo "memory.swap.free:int:$(( ${swap_total_bytes} - ${swap_used_bytes} ))"
