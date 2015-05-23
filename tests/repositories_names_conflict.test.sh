#!/bin/bash

set -u

TMPDIR=$(tests_tmpdir)

tests_do start_sould
tests_assert_success

tests_do create "repository_foo"
tests_assert_stdout_re "201 Created"

tests_do create "repository_foo"
tests_assert_stdout_re "409 Conflict"
