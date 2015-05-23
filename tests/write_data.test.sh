#!/bin/bash

set -u

TMPDIR=$(tests_tmpdir)

tests_do start_sould
tests_assert_success

tests_do create "repository_foo"

FOO_FILE="$TMPDIR/file_foo"
BAR_FILE="$TMPDIR/dir_bar/file_bar"

FOO_CONTENT="Foo content"
BAR_CONTENT="some bar data"

tests_do mkdir $TMPDIR/dir_bar

tests_do echo "$FOO_CONTENT" "> $FOO_FILE"
tests_do echo "$BAR_CONTENT" "> $BAR_FILE"

tests_do update repository_foo $FOO_FILE $BAR_FILE
tests_assert_stdout_re "200 OK"

tests_diff "$FOO_CONTENT" "$TMPDIR/repository_foo/$FOO_FILE"
tests_diff "$BAR_CONTENT" "$TMPDIR/repository_foo/$BAR_FILE"
