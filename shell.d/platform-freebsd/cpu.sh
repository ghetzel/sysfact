#!/bin/sh

echo "cpu.count:int:$(sysctl -n hw.ncpu)"
echo "cpu.model:str:$(sysctl -n hw.model)"
