#!/bin/bash

set -u

master="0"
slave_pa="1"
slave_re="2"

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
tests_do run_sould $config_slave_pa true
tests_assert_success

tests_do run_sould $config_slave_re true
tests_assert_success

tests_do run_sould $config_master true
tests_assert_success

# create upstream git repository with one commit
tests_do create_repository "upstream"
tests_assert_success

tests_do create_commit "upstream" "foo"
tests_assert_success

mirror_name="mirror/for/upstream"

# send pull request master, he must propagate request to slaves
tests_do request_pull $master $mirror_name `tests_tmpdir`/upstream
# should be '201 Created' instead of '200 OK' because master does not have
# mirror to repository 'upstream' yet
tests_assert_stderr_re "HTTP/1.1 201 Created"

# check for successfully propagate request to pulls, check commit in all
# storages
storages=("$storage_slave_re" "$storage_slave_pa" "$storage_master")
for storage in $storages; do
    mirror_dir=$storage/$mirror_name
    tests_test -d $mirror_dir

    tests_cd $mirror_dir
    tests_do git log
    tests_assert_success
    tests_assert_stdout_re "test-foo-commit"
done

tests_tmp_cd upstream
tests_assert_success

tests_do create_commit "upstream" "bar"
tests_assert_success

tests_do request_pull $master $mirror_name `tests_tmpdir`/upstream
tests_assert_stderr_re 'HTTP/1.1 200 OK'

for storage in $storages; do
    mirror_dir=$storage/$mirror_name
    tests_cd $mirror_dir

    tests_do git log
    tests_assert_stdout_re "test-foo-commit"
    tests_assert_stdout_re "test-bar-commit"
done

# corrupting 'pa' slave storage
tests_do mv $storage_slave_pa `tests_tmpdir`/backup_storage_pa
tests_assert_success

tests_do ln -sf /dev/null $storage_slave_pa
tests_assert_success

tests_do request_pull $master $mirror_name `tests_tmpdir`/upstream
tests_assert_stderr_re 'HTTP/1.1 502 Bad Gateway'
# should show message from 'pa' slave
tests_assert_stdout_re "http status is '500 Internal Server Error'"
tests_assert_stdout_re "$storage_slave_pa/$mirror_name: not a directory"

# corrupting upstream
tests_do mv `tests_tmpdir`/upstream `tests_tmpdir`/backup_upstream
tests_assert_success

tests_do request_pull $master $mirror_name `tests_tmpdir`/upstream
# master should show all messages, which returned by slaves
tests_assert_stderr_re 'HTTP/1.1 503 Service Unavailable'
tests_assert_stdout_re "slave '$addr_slave_pa'.*500 Internal Server Error"
tests_assert_stdout_re "slave '$addr_slave_re'.*500 Internal Server Error"
tests_assert_stdout_re "$storage_slave_pa/$mirror_name: not a directory"

#restoring sould upstream and storage_pa directory
tests_do mv `tests_tmpdir`/backup_upstream `tests_tmpdir`/upstream
tests_assert_success

tests_do rm -f $storage_slave_pa
tests_assert_success
tests_do mv `tests_tmpdir`/backup_storage_pa $storage_slave_pa
tests_assert_success

# does cluster restored?
tests_do request_pull $master $mirror_name `tests_tmpdir`/upstream
tests_assert_stderr_re 'HTTP/1.1 200 OK'
