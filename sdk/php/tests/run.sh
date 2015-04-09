#!/bin/bash

# var : process count
# var : value bytes 
BinDir=`dirname $0`
ProjectHome="$BinDir"
echo "The project dir:  "$ProjectHome

mkdir -p /tmp/shadow/


size=$1
concurrency=$2
duration=$3
read=$4
write=$5




echo "start..."
cd $ProjectHome
for ((i=1;i<=$concurrency;i++))
do
    echo "start $i php process"
    php benchmark.php $size $i $duration $read $write >/tmp/shadow/$i.log  &
done

echo "boot in background."
exit 0
