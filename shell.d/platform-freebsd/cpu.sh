#!/bin/sh

cpu_count="$(sysctl -n hw.ncpu)"

echo "cpu.count:int:${cpu_count}"
echo "cpu.model:str:$(sysctl -n hw.model)"

for i in $(seq 0 $(( ${cpu_count} - 1 ))); do
    cpu_n_curtemp="$(sysctl -n dev.cpu.${i}.temperature 2>/dev/null | awk '{print $1 }' | tr -d 'C')"
    cpu_n_maxtemp="$(sysctl -n dev.cpu.${i}.coretemp.tjmax 2>/dev/null | awk '{print $1 }' | tr -d 'C')"

    if [ -n "${cpu_n_curtemp}" ]; then
        echo "cpu.cores.${i}.temperature:float:${cpu_n_curtemp}"

    fi

    if [ -n "${cpu_n_maxtemp}" ]; then
        echo "cpu.cores.${i}.max_temperature:float:${cpu_n_maxtemp}"
    fi
done
