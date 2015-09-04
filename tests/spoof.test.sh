#!/bin/bash

set -u

master="0"
slave_pa="1"

addr_slave_pa=`get_listen_addr $slave_pa`
storage_slave_pa="`get_storage $slave_pa`"
config_slave_pa="`get_config_slave $addr_slave_pa $storage_slave_pa`"

slaves="$addr_slave_pa"

addr_master="`get_listen_addr $master`"
storage_master="`get_storage $master`"
config_master="`get_config_master $addr_master $storage_master 3000 $slaves`"

# run all servers
tests_ensure run_sould $config_slave_pa true
tests_ensure run_sould $config_master true

# create upstream git repository with one commit
tests_ensure create_repository "upstream"

tests_ensure create_commit "upstream" "commit_A"
tests_ensure create_commit "upstream" "commit_B"

tests_tmp_cd "upstream"
tests_ensure git tag "tag-before-c"
tests_ensure create_commit "upstream" "commit_C"
tests_ensure git tag "tag-after-c"
tests_ensure git reset --hard tag-before-c
tests_ensure git tag -d tag-before-c

tests_ensure request_pull_spoof \
    $master "mirror/for/upstream" `tests_tmpdir`/upstream master tag-after-c

tests_cd $storage_slave_pa/mirror/for/upstream
tests_ensure git log
tests_assert_stdout_re "commit_A"
tests_assert_stdout_re "commit_B"
tests_assert_stdout_re "commit_C"

tests_cd $storage_master/mirror/for/upstream
tests_ensure git log
tests_assert_stdout_re "commit_A"
tests_assert_stdout_re "commit_B"
tests_assert_stdout_re "commit_C"
