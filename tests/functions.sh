#!/bin/bash

# Function 'get_listen_addr' returns listen address for sould.
# Args:
#   $1 - number of running sould
get_listen_addr() {
    local number=$1

    echo localhost:$((60000+$number))
}

# Function 'get_storage' returns storage temporary directory path.
# Args:
#   $1 - number of running sould
get_storage() {
    local number="$1"

    local directory=`tests_tmpdir`/storage/$number

    tests_do mkdir -p $directory
    tests_assert_success

    echo $directory
}

# Function 'get_config_slave' returns basic configuration for sould in
# non-master mode.
# Args:
#    $1 - listening address
#    $2 - storage directory path
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

# Function 'get_config_master' returns config for sould in master mode.
# master config inherits slave config.
# Args
#   $1 - listening address
#   $2 - storage directory path
#   $@ - slaves
get_config_master() {
    local listen="$1"
    local storage="$2"
    local timeout="$3"
    shift 3

    local slaves='"'$(sed 's/ /", "/g' <<< "$@")'"'

    local path=`get_config_slave $listen $storage`

    local config="
master = true
timeout = $timeout
slaves = [$slaves]
"

    echo "$config" >> "$path"
    echo "$path"
}

# Function 'run_sould' starts background work for sould with specified config.
# Returns unique background work identifier.
# Args:
#    $1 - configuration file
#    $2 - unsecure mode (boolean value)
run_sould() {
    local config="$1"
    local unsecure=$2

    local params="-c $1"
    if $unsecure; then
        params="$params --unsecure"
    fi

    local listen=`cat $config | awk '/listen/{print $3}' | sed 's/"//g'`

    tests_debug "running sould server on $listen"

    local bg_id=`tests_background "$SOULD_BIN $params"`
    local bg_pid=`tests_background_pid $bg_id`


    # 10 seconds
    local check_max=100
    local check_counter=0

    while true; do
        sleep 0.1

        if ! kill -0 $bg_pid; then
            tests_debug "sould has gone away..."
            return 1
        fi

        grep -q "$listen" <<< "`netstat -tl`"
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

# Function 'requests_pull' do POST request to a sould server, which should be
# specified by number.
# Args:
#    $1 - sould server number
#    $2 - mirror name
#    $3 - mirror origin (clone url)
request_pull() {
    local number="$1"
    local name="$2"
    local origin="$3"

    curl -s -v -X POST \
        -m 10 \
        --data "name=$name&origin=$origin" \
        `get_listen_addr $number`/
}

# Function 'request_tar' do GET request to a sould server, which should be
# specified by number, tar archive content will be available in stdout.
# Args:
#    $1 - sould server number
#    $2 - mirror name
request_tar() {
    local number="$1"
    local mirror="$2"

    curl -s -v -X GET \
        -m 10 \
        `get_listen_addr $number`/$mirror
}

# Function 'create_git' creates a git commit in directory with git repository,
# which should be created by function 'create_repository'.
# Function creates file with specified name and adds commit with content
# 'test-$file-commit'
# Args:
#   $1 - repository directory (relative path to temporary directory)
#   $2 - file
create_commit() {
    local repository="$1"
    local file="$2"

    tests_tmp_cd $repository
    tests_do touch $file
    tests_assert_success
    tests_do git add $file
    tests_assert_success
    tests_do git commit -m test-$file-commit
    tests_assert_success
}

# Function 'create_repository' creates a git repository in temporary test
# directory.
# Args:
#   $1 - directory name
create_repository() {
    local name="$1"

    tests_mkdir $name
    tests_tmp_cd $name

    tests_do git init
    tests_assert_success
}
