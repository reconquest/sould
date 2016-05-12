#!/bin/bash

set -e

cd $(readlink -f  $(dirname $0))

export BUILD=$(mktemp -u -t buildXXXXXX)
go build -o $BUILD ../
trap "rm $BUILD" EXIT

# bash tests library
if [ ! -f lib/tests.sh ]; then
    echo "'tests.sh' dependency is missing"
    echo "trying fix this via updating git submodules"
    git submodule init
    git submodule update

    if [ ! -f lib/tests.sh ]; then
        echo "file 'lib/tests.sh' not found"
        exit 1
    fi
fi


./lib/tests.sh -d testcases -s util/setup.sh -Aa $@
