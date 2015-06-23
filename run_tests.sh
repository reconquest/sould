#!/bin/bash

#set -x

SOULD_LISTEN="localhost:60088"
SOULD_BIN="$(readlink -f sould)"

go build -o $SOULD_BIN
if [ $? -ne 0 ]; then
    echo "can't build project"
    exit 1
fi

source tests/functions.sh

# bash tests library
if [ ! -f tests/lib/tests.sh ]; then
    echo "'tests.sh' dependency is missing"
    echo "trying fix this via updating git submodules"
    git submodule init
    git submodule update

    if [ ! -f tests/lib/tests.sh ]; then
        echo "file 'tests/lib/tests.sh' not found"
        exit 1
    fi
fi

source tests/lib/tests.sh

TEST_VERBOSE=10

cd tests/
tests_run_all
