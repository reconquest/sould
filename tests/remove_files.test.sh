#!/bin/bash

set -u

TMPDIR=$(tests_tmpdir)

tests_do start_sould
tests_assert_success

tests_do create "trash_repo"
tests_assert_success "201 Created"

tests_do mkdir $TMPDIR/recycler_dir/

tests_do touch "$TMPDIR/recycler_dir/stable_file"

tests_do touch "$TMPDIR/recycler_dir/removing_file_foooo"
tests_do touch "$TMPDIR/recycler_dir/removing_file_bar"

tests_cd $TMPDIR

tests_do update "trash_repo" \
    recycler_dir/removing_file_foooo \
    recycler_dir/removing_file_bar \
    recycler_dir/stable_file
tests_assert_stdout_re "200 OK"

tests_cd $OLDPWD

RM_QUERY="op=remove"
RM_QUERY="$RM_QUERY&files[]=recycler_dir/removing_file_foooo"
RM_QUERY="$RM_QUERY&files[]=recycler_dir/removing_file_bar"

tests_do curl -v -s --data "$RM_QUERY" -X PUT \
    http://$SOULD_LISTEN/trash_repo '2>&1'

tests_assert_stdout_re "200 OK"

tests_do find $TMPDIR/recycler_dir/

tests_test -f $TMPDIR/recycler_dir/stable_file

tests_test ! -f $TMPDIR/recycler_dir/removing_file_foooo
tests_test ! -f $TMPDIR/recycler_dir/removing_file_bar
