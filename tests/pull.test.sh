#!/bin/bash

set -u

master="0-master"
slave_pa="1-slave_pa"
slave_re="2-slave_re"

addr_slave_pa=`get_listen_addr $slave_pa`
storage_slave_pa="`get_storage $slave_pa`"
config_slave_pa="`get_config_slave $addr_slave_pa $storage_slave_pa`"

addr_slave_re=`get_listen_addr $slave_re`
storage_slave_re="`get_storage $slave_re`"
config_slave_re="`get_config_slave $addr_slave_re $storage_slave_re`"

slaves="$addr_slave_pa $addr_slave_re"

addr_master="`get_listen_addr $master`"
storage_master="`get_storage $master`"
config_master="`get_config_master $addr_master $storage_master 3000 $slaves`"

# run all servers
tests_ensure run_sould $config_slave_pa true
tests_ensure run_sould $config_slave_re true
tests_ensure run_sould $config_master true

# create upstream git repository with one commit
tests_ensure create_repository "upstream"

tests_ensure create_commit "upstream" "foo"

mirror_name="mirror/for/upstream"

# send pull request master, he must propagate request to slaves
tests_do request_pull $master $mirror_name `tests_tmpdir`/upstream
tests_assert_stdout_re "200 OK"

# check for successfully propagating request to slaves, check commit in all
# storages
storages=("$storage_slave_re" "$storage_slave_pa" "$storage_master")
for storage in $storages; do
    mirror_dir=$storage/$mirror_name
    tests_test -d $mirror_dir

    tests_cd $mirror_dir
    tests_ensure git log
    tests_assert_stdout_re "test-foo-commit"
done

tests_tmp_cd upstream
tests_assert_success

tests_ensure create_commit "upstream" "bar"

tests_ensure request_pull $master $mirror_name `tests_tmpdir`/upstream
tests_assert_stdout_re '200 OK'

for storage in $storages; do
    mirror_dir=$storage/$mirror_name
    tests_cd $mirror_dir

    tests_do git log
    tests_assert_stdout_re "test-foo-commit"
    tests_assert_stdout_re "test-bar-commit"
done

# corrupting 'pa' slave storage
tests_ensure mv $storage_slave_pa `tests_tmpdir`/backup_storage_pa

tests_ensure ln -sf /dev/null $storage_slave_pa

tests_do request_pull $master $mirror_name `tests_tmpdir`/upstream
tests_assert_stdout_re '< HTTP/1.1 502 Bad Gateway'
# should show message from 'pa' slave
tests_assert_stdout_re "^[^<].*500 Internal Server Error"
tests_assert_stdout_re "$storage_slave_pa/$mirror_name: not a directory"

# corrupting upstream
tests_ensure mv `tests_tmpdir`/upstream `tests_tmpdir`/backup_upstream

tests_do request_pull $master $mirror_name `tests_tmpdir`/upstream
tests_assert_stdout_re '< HTTP/1.1 503 Service Unavailable'

tests_ensure grep -A 5 -P "slave $addr_slave_pa" `tests_tmpdir`/response
tests_assert_stdout_re "500 Internal Server Error"
tests_assert_stdout_re "$storage_slave_pa/$mirror_name: not a directory"

tests_ensure grep -A 1 -P "slave $addr_slave_re" `tests_tmpdir`/response
tests_assert_stdout_re "500 Internal Server Error"

#restoring sould upstream and storage_pa directory
tests_ensure mv `tests_tmpdir`/backup_upstream `tests_tmpdir`/upstream

tests_ensure rm -f $storage_slave_pa
tests_ensure mv `tests_tmpdir`/backup_storage_pa $storage_slave_pa

# does cluster restored?
tests_do request_pull $master $mirror_name `tests_tmpdir`/upstream
tests_assert_stdout_re '200 OK'
