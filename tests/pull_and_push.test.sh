#!/bin/bash

set -u

TMPDIR=$(tests_tmpdir)

tests_do start_sould
tests_assert_success

tests_do create "local_repository"
tests_assert_stdout_re "201 Created"

FOO_FILE="$TMPDIR/foo_file"
FOO_CONTENT="foo data"
tests_do echo "$FOO_CONTENT" "> $FOO_FILE"

tests_do update "local_repository" "$FOO_FILE"
tests_assert_stdout_re "200 OK"

# creating 'remote' bare repository from copy of 'update_testings' repository
tests_do cp -r $TMPDIR/local_repository/.git $TMPDIR/remote_git

tests_cd $TMPDIR/remote_git
tests_do git config --bool core.bare true

tests_cd $TMPDIR/local_repository
tests_do git remote add origin $TMPDIR/remote_git
tests_do git config push.default current
tests_do git fetch
tests_do git branch --set-upstream-to=origin/master master

# for testing pull writing something to remote repository
tests_do cp -r $TMPDIR/local_repository $TMPDIR/writing_local_repository

tests_cd $TMPDIR/writing_local_repository
tests_do echo "pull me please" "> pullme"
tests_do git add pullme
tests_do git commit -m '"pull test file added"'
tests_do git push -u

QUX_FILE="$TMPDIR/qux_file"
QUX_CONTENT="qux data"
tests_do echo "$QUX_CONTENT" "> $QUX_FILE"

tests_cd $TMPDIR
tests_do update "local_repository" "$QUX_FILE"
tests_assert_stdout_re "200 OK"

LOCAL_LOG=$(cd $TMPDIR/local_repository && git log)
REMOTE_LOG=$(cd $TMPDIR/remote_git && git log)

tests_diff "$LOCAL_LOG" "$REMOTE_LOG"
