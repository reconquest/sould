#!/bin/bash

get_listen_addr() {
    local number=$1

    echo "localhost:"$((60000+$number))
}

get_storage() {
    local number="$1"

    local directory="$(tests_tmpdir)/storage/$number"

    tests_do mkdir -p $directory
    tests_assert_success

    echo $directory
}

get_config_slave() {
    local listen="$1"
    local storage="$2"

    local path="$storage/.config"

    local config="
listen = \"$listen\"
storage = \"$storage\"
"

    echo "$config" > "$path"
    echo "$path"
}

get_config_master() {
    local listen="$1"
    local storage="$2"
    local timeout="$3"
    shift 3

    local slaves='"'$(sed 's/ /", "/g' <<< "$@")'"'

    local path="$(get_config_slave $listen $storage)"

    local config="
master = true
timeout = $timeout
slaves = [$slaves]
"

    echo "$config" >> "$path"
    echo "$path"
}

run_sould() {
    local config="$1"
    local unsecure=$2

    local params="-c $1"
    if $unsecure; then
        params="$params --unsecure"
    fi

    local listen="$(cat $config | awk '/listen/{print $3}' | sed 's/"//g')"

    tests_debug "running sould server on $listen"

    local bg_id=$(tests_background "$SOULD_BIN $params")
    local bg_pid=$(tests_background_pid $bg_id)


    # 10 seconds
    local check_max=100
    local check_counter=0

    while true; do
        sleep 0.1

        if ! kill -0 $bg_pid; then
            tests_debug "sould has gone away..."
            return 1
        fi

        grep -q "$listen" <<< "$(netstat -tl)"
        local grep_result=$?
        if [ $grep_result -eq 0 ]; then
            break
        fi

        check_counter=$(($check_counter+1))
        if [ $check_counter -ge $check_max ]; then
            tests_debug "sould not started listening on $listen"
            return 1
        fi
    done

    echo $bg_id
}

request_pull() {
    local number="$1"
    local name="$2"
    local origin="$3"

    curl -s -v -X POST \
        -m 10 \
        --data "name=$name&origin=$origin" \
        "$(get_listen_addr $number)" 2>&1
}

# update() function updates repository with specified name and uploads files
# Args:
# $1 - mirror name
# $@ - files
request_update() {
    local name="$1"
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
