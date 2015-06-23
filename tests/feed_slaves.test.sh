#!/bin/bash

set -u

addr_slave_pa=`get_listen_addr 1`
storage_slave_pa="`get_storage 1`"
config_slave_pa="`get_config_slave $addr_slave_pa $storage_slave_pa`"

addr_slave_re=`get_listen_addr 2`
storage_slave_re="`get_storage 2`"
config_slave_re="`get_config_slave $addr_slave_re $storage_slave_re`"

slaves="$addr_slave_pa $addr_slave_re"

addr_master="`get_listen_addr 0`"
storage_master="`get_storage 0`"
config_master="`get_config_master $addr_master $storage_master 3000 $slaves`"

tests_do run_sould $config_slave_pa true
tests_assert_success

tests_do run_sould $config_slave_re true
tests_assert_success

tests_do run_sould $config_master true
tests_assert_success

tests_mkdir upstream

tests_tmp_cd upstream

tests_do git init
tests_do touch foo
tests_do git add foo
tests_do git commit -m fooed

mirror_name="mirror/for/upstream"

# feed master (server 0), he must feed slaves
tests_do request_pull 0 $mirror_name `tests_tmpdir`/upstream
tests_assert_stdout_re '201 Created'

storages=("$storage_slave_re" "$storage_slave_pa" "$storage_master")
for storage in $storages; do
    mirror_dir=$storage/$mirror_name
    tests_test -d $mirror_dir

    tests_cd $mirror_dir
    tests_do git log
    tests_assert_stdout_re 'fooed'
done

tests_tmp_cd upstream
tests_do touch bar
tests_do git add bar
tests_do git commit -m bared

tests_do request_pull 0 $mirror_name `tests_tmpdir`/upstream
tests_assert_stdout_re '200 OK'

for storage in $storages; do
    mirror_dir=$storage/$mirror_name
    tests_cd $mirror_dir

    tests_do git log
    tests_assert_stdout_re 'fooed'
    tests_assert_stdout_re 'bared'
done

tests_do
