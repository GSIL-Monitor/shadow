#!/bin/bash
export env=prod

BinDir=`dirname $0`
ProjectHome="$BinDir/../"
ProjectHome=`readlink -f $ProjectHome`


go_path=`go env GOPATH`
app_path=$go_path/bin/
cd $ProjectHome
mkdir -p /var/run/shadow/
# run in daemon
$app_path/agent -cnf_basedir=$go_path/src/shadow/agent/cnf/ >/dev/null 2>&1    &
echo "$!" > /var/run/shadow/agent.pid 
