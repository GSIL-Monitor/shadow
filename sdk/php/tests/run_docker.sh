#!/bin/bash

# var : value bytes

BinDir=`dirname $0`
ProjectHome="$BinDir/../"

mkdir -p /tmp/docker/

# KB
size=$1

echo "start"

for ((i=1;i<=1;i++))
do
        php docker_test.php $size 1 >/tmp/docker/$size"_KB_docker".$i.log &
        php docker_test.php $size 0 >/tmp/docker/$size"_KB_non-docker".$i.log &
done


echo "boot in background"
exit 0