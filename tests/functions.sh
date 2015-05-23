#!/bin/bash

start_sould() {
    local stdout=$(mktemp)
    local TMPDIR=$(tests_tmpdir)

    tests_background "$SOULD_BIN -l $SOULD_LISTEN -d $TMPDIR"
    local bg_pid=$(tests_background_pid)

    # 10 seconds
    local check_max=100
    local check_counter=0

    while true; do
        sleep 0.1

        if ! kill -0 $bg_pid; then
            tests_debug "sould has gone away..."
            return 1
        fi

        grep -q "$SOULD_LISTEN" <<< "$(netstat -tl)"
        local grep_result=$?
        if [ $grep_result -eq 0 ]; then
            return 0
        fi

        check_counter=$(($check_counter+1))
        if [ $check_counter -ge $check_max ]; then
            tests_debug "sould not started listening on $SOULD_LISTEN"
            return 1
        fi
    done
}

# create() function creates new repository with specified name
create() {
    local name=$1
    curl -s -v -X POST \
        --data "name=$name" \
        $SOULD_LISTEN 2>&1
}

# update() function updates repository with specified name and uploads files
update() {
    local name=$1
    shift

    local files=""
    while (( "$#" )); do
        files="$files -F files[]=@$1;filename=$1"
        shift
    done

    curl -s -v -X PUT \
        -F "op=update" \
        $files \
        http://$SOULD_LISTEN/$name 2>&1
}
