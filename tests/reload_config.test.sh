#!/bin/bash

set -u

sighup() {
    kill -HUP $pid;
}

addr=`get_listen_addr 0`
storage=`get_storage 0`
config="`get_config_slave $addr $storage`"

sleep_max=10

tests_ensure run_sould "$config" true

job_id=$(cat `tests_stdout`)
pid=`tests_background_pid $job_id`
stderr="`tests_background_stderr $job_id`"

tests_wait_file_changes sighup $stderr 0.1 10
tests_assert_re "$stderr" "reloaded"

tests_do rm -rf $config

tests_wait_file_changes sighup $stderr 0.1 10

tests_assert_re "$stderr" "can't reload config"
tests_assert_re "$stderr" "no such file or directory"
