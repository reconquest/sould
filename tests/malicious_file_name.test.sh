#!/bin/bash

set -u

TMPDIR=$(tests_tmpdir)

tests_do start_sould
tests_assert_success

tests_do mkdir -p $TMPDIR/malicious/.git
tests_do echo "hook" "> $TMPDIR/malicious/.git/hook"

tests_do create "update_testings"
tests_assert_stdout "201 Created"
tests_assert_success

tests_do update "update_testings" "$TMPDIR/malicious/.git/hook"
tests_assert_stdout_re "403 Forbidden"
