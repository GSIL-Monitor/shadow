#!/bin/sh

#############################################################
# monitor server running									#
# restart server if detect it is not running				#
# monitor.sh must be in the same directory with the server	#
#############################################################
shell_dir=$(dirname $0)
shell_dir=${shell_dir/\./$(pwd)}
log_dir=/data/logs/shadow/
run_dir=/var/run/shadow/

function monitor_pid() {
    mkdir -p $log_dir 
    mkdir -p $run_dir

    cd $shell_dir

    pid=`cat /var/run/shadow/agent.pid`
    process_count=`ps aux| grep $1 | grep ' '$pid' ' | grep -v shadow_monitor.sh | grep -v grep| wc -l`
    if [ $process_count == 0 ]; then
        date >> $log_dir/restart.log
        echo "server stopped, process_cnt=$process_count" >> $log_dir/restart.log

        ./daemon_handle ./$1 -cnf_basedir=/var/workspace-go/src/shadow/agent/cnf/
    fi
}

function monitor() {
    cd $shell_dir

    while true
    do
        monitor_pid $1
        sleep 5
    done
}

case $1 in
	agent)
		monitor $1
		;;
	*)
		echo "Usage: "
		echo "  ./shadow_monitor.sh (agent)"
		;;
esac
