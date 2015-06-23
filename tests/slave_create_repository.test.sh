#!/bin/bash

set -u

addr="`get_listen_addr 0`"
storage="`get_storage 0`"
config="`get_config_slave $addr $storage`"

tests_do run_sould $config true
tests_assert_success

tests_mkdir upstream

tests_tmp_cd upstream

tests_do git init
tests_do touch foo
tests_do git add foo
tests_do git commit -m fooed

mirror_name="mirror/for/upstream"
mirror_dir=$storage/$mirror_name

tests_do request_pull 0 $mirror_name `tests_tmpdir`/upstream
tests_assert_stdout_re '201 Created'

tests_test -d $mirror_dir

tests_cd $mirror_dir
tests_do git log
tests_assert_stdout_re 'fooed'

tests_tmp_cd upstream
tests_do touch bar
tests_do git add bar
tests_do git commit -m bared

tests_do request_pull 0 $mirror_name `tests_tmpdir`/upstream
tests_assert_stdout_re '200 OK'

tests_cd $mirror_dir
tests_do git log
tests_assert_stdout_re 'fooed'
tests_assert_stdout_re 'bared'
