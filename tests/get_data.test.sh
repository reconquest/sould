#!/bin/bash

set -u

TMPDIR=$(tests_tmpdir)

tests_do start_sould
tests_assert_success

REPO_NAME="repository_foo"

tests_do create $REPO_NAME
tests_assert_stdout_re "201 Created"

tests_test -d $TMPDIR/$REPO_NAME

tests_test -d $TMPDIR/$REPO_NAME/.git

FILE_NAME="file_for_get_test"
FILE_CONTENT="data for get test"
FILE="$TMPDIR/$REPO_NAME/$FILE_NAME"

tests_do mkdir -p $(dirname $FILE)
tests_do echo "$FILE_CONTENT" "> $FILE"

tests_cd $TMPDIR/$REPO_NAME

tests_do update $REPO_NAME $FILE_NAME
tests_assert_stdout_re "200 OK"

tests_cd $OLDPWD

# tar archive
tests_do "curl -s http://$SOULD_LISTEN/$REPO_NAME > $TMPDIR/actual_archive.tar"
tests_assert_success

tests_cd $TMPDIR/$REPO_NAME
tests_do git archive -o expected_archive.tar HEAD
tests_cd $OLDPWD

EXPECTED_TREE="$(tar tf $TMPDIR/$REPO_NAME/expected_archive.tar)"

tests_do tar tf $TMPDIR/actual_archive.tar
tests_assert_success

tests_diff "$EXPECTED_TREE" stdout

# test case for one file
tests_do curl -s http://$SOULD_LISTEN/$REPO_NAME/$FILE_NAME
tests_diff "$FILE_CONTENT" stdout
