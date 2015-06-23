#!/bin/bash

set -u

config="$(get_config_slave $(get_listen_addr 0) $(get_storage 0))"

tests_do run_sould "$config" true
tests_assert_success

bg_id=$(cat $TEST_STDOUT)
pid=$(tests_background_pid $bg_id)
stderr="$(tests_background_stderr $bg_id)"

kill -HUP $pid
tests_assert_re "$stderr" "reloaded"

rm -rf $config

kill -HUP $pid
tests_assert_re "$stderr" "can't reload config"
tests_assert_re "$stderr" "no such file or directory"
