#!/bin/bash

addr=`get_listen_addr 0`
storage=`get_storage 0`
config=`get_config_slave $addr $storage`

tests_do run_sould "$config" true
tests_assert_success

tests_do create_repository "upstream"
tests_assert_success

tests_do create_commit "upstream" "file_foo"
tests_assert_success

mirror_name="mirror/of/upstream"

tests_do request_tar 0 $mirror_name
tests_assert_stderr_re '404 Not Found'

tests_do request_pull 0 $mirror_name `tests_tmpdir`/upstream
tests_assert_stderr_re '201 Created'

tests_do request_tar 0 $mirror_name '>' `tests_tmpdir`/archive.tar
tests_assert_stderr_re '200 OK'

tests_do tar -xlvf `tests_tmpdir`/archive.tar
tests_assert_success
tests_assert_stdout_re 'file_foo'

tests_do mv `tests_tmpdir`/upstream `tests_tmpdir`/backup_upstream
tests_assert_success

