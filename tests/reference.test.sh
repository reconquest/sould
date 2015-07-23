#!/bin/bash

set -u

mirror_name="mirror/of/upstream"

addr=`get_listen_addr 0`
storage=`get_storage 0`
config=`get_config_slave $addr $storage`

tests_ensure run_sould "$config" true

tests_ensure create_repository "upstream"

tests_do create_commit "upstream" "file_foo"
tests_assert_success

tests_tmp_cd "upstream"
commit_foo="$(git rev-parse HEAD)"

tests_ensure request_pull 0 $mirror_name `tests_tmpdir`/upstream
tests_ensure request_tar 0 $mirror_name '>' `tests_tmpdir`/archive.tar

tests_ensure tar -xlvf `tests_tmpdir`/archive.tar
tests_assert_stdout_re "file_foo"

tests_ensure create_commit "upstream" "file_bar"

tests_ensure request_pull 0 $mirror_name `tests_tmpdir`/upstream
tests_ensure request_tar 0 $mirror_name '>' `tests_tmpdir`/archive.tar

tests_ensure tar -xlvf `tests_tmpdir`/archive.tar
tests_assert_stdout_re "file_foo"
tests_assert_stdout_re "file_bar"

tests_ensure \
    request_tar 0 $mirror_name "$commit_foo" '>' `tests_tmpdir`/archive.tar

tests_ensure tar -xlvf `tests_tmpdir`/archive.tar
tests_assert_stdout_re "file_foo"
tests_do tests_re "stdout" "file_bar"
tests_assert_exitcode 1
