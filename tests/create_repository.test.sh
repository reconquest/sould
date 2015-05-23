#!/bin/bash

set -u

TMPDIR=$(tests_tmpdir)

# wrap 'PATH' env variable for 'sould-init'
OLDPATH=$PATH
PATH=$TMPDIR:$PATH
tests_debug "PATH variable value is set to $PATH"

tests_do start_sould
tests_assert_success

# Create sould backend init script, which must receive repository name and
# must have repository directory as working directory
cat > $TMPDIR/sould-init <<CODE
#!/bin/bash

NAME=\$1
PWD=\$(pwd)
echo "name = \$NAME | pwd = \$PWD" > $TMPDIR/check_init_script
CODE

tests_debug "created sould init file"
tests_do chmod +x $TMPDIR/sould-init

TEST_NAMES=(
    "repo_foo"
    "repo_bar"
    "repo_pool_foo/repo_pool_bar/repo_qux"
)

for REPO_NAME in ${TEST_NAMES[*]}; do
    tests_do create "$REPO_NAME"
    tests_assert_stdout_re "201 Created"

    tests_test -d "$TMPDIR/$REPO_NAME"

    tests_test -d "$TMPDIR/$REPO_NAME/.git"

    tests_test -f "$TMPDIR/check_init_script"

    tests_diff \
        "name = $REPO_NAME | pwd = $TMPDIR/$REPO_NAME" \
        "$TMPDIR/check_init_script"
done

PATH=$OLDPATH
tests_do "PATH variable value restored"
