#!/bin/bash

set -u

sould="0"
addr="`get_listen_addr $sould`"
storage="`get_storage $sould`"
config="`get_config_slave $addr $storage localhost:9418`"

tests_ensure run_sould $config true

tests_ensure create_repository "upstream"
tests_ensure create_commit "upstream" "foo"

mirror_name="pool/mirror"

tests_ensure request_pull $sould $mirror_name `tests_tmpdir`/upstream
tests_assert_stderr_re "201 Created"

git_cmd="git daemon --base-path=$storage --reuseaddr --port=9419 --export-all"
tests_background "$git_cmd"

tests_ensure git clone git://localhost/pool/mirror `tests_tmpdir`/cloned
tests_tmp_cd cloned
tests_do git log
tests_assert_stdout_re 'foo'
tests_tmp_cd /

tests_ensure rm -rf `tests_tmpdir`/cloned

tests_ensure mv `tests_tmpdir`/upstream `tests_tmpdir`/backup_upstream

tests_ensure git clone git://localhost/pool/mirror `tests_tmpdir`/cloned
tests_tmp_cd cloned
tests_do git log
tests_assert_stdout_re 'foo'
tests_tmp_cd /

tests_ensure rm -rf `tests_tmpdir`/cloned

tests_do request_pull $sould $mirror_name `tests_tmpdir`/upstream
tests_assert_stderr_re '500 Internal Server Error'

tests_do git clone git://localhost/pool/mirror `tests_tmpdir`/cloned
tests_assert_exitcode 128

tests_ensure mv `tests_tmpdir`/backup_upstream `tests_tmpdir`/upstream

tests_ensure git clone git://localhost/pool/mirror `tests_tmpdir`/cloned
tests_tmp_cd cloned
tests_do git log
tests_assert_stdout_re 'foo'
