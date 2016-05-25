#!/bin/bash

shopt -s expand_aliases

alias @var='tests:value'

tests:make-tmp-dir storage
tests:make-tmp-dir port
tests:make-tmp-dir task
tests:make-tmp-dir config

:get-port() {
    local identifier="$@"

    if [ -f "port/$identifier" ]; then
        cat "port/$identifier"
        return
    fi

    local number=$((10000+$RANDOM))

    tests:put-string "port/$identifier" "$number"

    echo "$number"
}

:get-storage() {
    local identifier="$1"

    tests:make-tmp-dir "storage/$identifier"
    echo "storage/$identifier"
}

:get-config() {
    local identifier="$1"

    echo "config/$identifier"
}

:configure-slave() {
    local identifier="$1"

    local storage="$(:get-storage "$identifier")"
    local http_port="$(:get-port "$identifier")"
    local git_port="$(:get-port "git_$identifier")"

    local config="$(:get-config "$identifier")"

    tests:put "$config" <<CONFIG
storage = "$storage"
[http]
    listen = "$(hostname):$http_port"
[git]
    listen = "$(hostname):$git_port"
    daemon = "$(hostname):9419"
CONFIG
}

:configure-master() {
    local identifier="$1"
    local timeout="$2"
    shift 2

    local slaves=()
    while [[ $# -gt 0 ]]; do
        slaves+=('"'$(hostname):$(:get-port "$1")'"')
        shift
    done

    if [[ -v slaves ]]; then
        slaves="${slaves[@]}"
        slaves=${slaves// /, }
    else
        slaves=""
    fi

    @var storage :get-storage "$identifier"
    @var config :get-config "$identifier"

    :configure-slave "$identifier"

    tests:put "$config" <<CONFIG
master = true
timeout = $timeout
slaves = [$slaves]
$(cat $config)
CONFIG
}

:sould-process() {
    local identifier="$1"

    @var task cat "task/$identifier"
    tests:get-background-pid "$task"
}

:sould-stdout() {
    local identifier="$1"

    @var task cat "task/$identifier"
    tests:get-background-stdout "$task"
}

:sould-stderr() {
    local identifier="$1"

    @var task cat "task/$identifier"
    tests:get-background-stderr "$task"
}

:sould-stop() {
    local identifier="$1"

    @var task cat "task/$identifier"
    tests:stop-background "$task"
}

:sould-start() {
    local identifier="$1"
    shift

    @var config :get-config "$identifier"
    @var port :get-port "$identifier"

    if [[ ! -f "$config" ]]; then
        :configure-slave "$identifier"
    fi

    tests:run-background task $BUILD -c $config ${@}
    tests:put-string "task/$identifier" "$task"

    @var sould_process tests:get-background-pid $task

    local i=0
    while :; do
        tests:describe "waiting for sould listening task"
        sleep 0.05

        local netstat=""
        if netstat=$(netstat -na | grep ":$port"); then
            tests:describe "$netstat"
            break
        fi

        i=$((i+1))
        if [ "$i" -gt 20 ]; then
            tests:fail "process doesn't started listening at $port"
        fi
    done
}

:request-pull() {
    local identifier="$1"
    local name="$2"
    local origin="$3"

    tests:pipe curl \
        -s \
        -v \
        -X POST \
        -m 10 \
        --data "name=$name&origin=$origin" \
        "$(hostname):$(:get-port $identifier)/" '2>&1'

    local exitcode=$(tests:get-exitcode)
    local stdout=$(tests:get-stdout-file)

    cp $stdout $(tests:get-tmp-dir)/response

    return $exitcode
}

:request-pull-with-spoof() {
    local identifier="$1"
    local name="$2"
    local origin="$3"
    local branch="$4"
    local tag="$5"

    tests:pipe curl -s -v -X POST \
        -m 10 \
        --data "name=$name&origin=$origin&spoof=1&branch=$branch&tag=$tag" \
        "$(hostname):$(:get-port $identifier)/"

    return $(tests:get-exitcode)
}

:request-tar() {
    local identifier="$1"
    local mirror="$2"

    local query=""
    if [ $# -gt 2 ]; then
        query="?ref=$3"
    fi

    tests:pipe curl -s -v -X GET \
        -m 10 \
        "$(hostname):$(:get-port $identifier)/$mirror$query"

    return $(tests:get-exitcode)
}

:git-repository() {
    local name="$1"

    tests:make-tmp-dir $name

    tests:cd-tmp-dir $name
    tests:ensure git init
    tests:cd
}

:git-commit() {
    local repository="$1"
    local file="$2"

    tests:cd-tmp-dir $repository
    tests:ensure touch $file
    tests:ensure git add $file
    tests:ensure git commit -m "$file"
    tests:cd
}

:git-tag() {
    local repository="$1"
    local name="$2"
    shift 2

    tests:cd-tmp-dir $repository
    tests:ensure git tag "$@" "$name"
    tests:cd
}

:git-reset() {
    local repository="$1"
    local treeish="$2"

    tests:cd-tmp-dir $repository
    tests:ensure git reset --hard "$treeish"
    tests:cd
}

:git-log() {
    local repository="$1"
    shift

    tests:cd-tmp-dir $repository
    tests:ensure git log "$@"
    tests:cd
}

:git-clone() {
    local identifier="$1"
    local name="$2"
    local to="$3"
    shift

    @var git_port :get-port git_$identifier
    tests:ensure git clone git://$(hostname):$git_port/$name $to
}

:git-clone-fail() {
    local identifier="$1"
    local name="$2"
    local to="$3"
    shift

    @var git_port :get-port git_$identifier
    tests:not tests:ensure git clone git://$(hostname):$git_port/$name $to
}

:git-server-start() {
    local identifier="$1"

    @var storage :get-storage sweety-slave

    tests:run-background task git daemon \
        --base-path=$storage --reuseaddr --port=9419  --export-all
}

:git-reference() {
    local repository="$1"

    tests:cd-tmp-dir $repository
    tests:pipe git rev-parse HEAD
    tests:assert-success
    tests:cd
}
