#!/bin/sh

echo "os.family:str:freebsd"
echo "os.distribution:str:$(uname -s)"
echo "os.version:str:$(uname -r)"
echo "os.description:str:$(uname -s) $(uname -r) $(uname -i)"
