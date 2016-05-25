#!/bin/bash

:sould-start little-slave --insecure

:git-repository upstream
:git-commit upstream file_foo

tests:ensure \
    :request-pull little-slave mir/ror $(tests:get-tmp-dir)/upstream
tests:ensure \
    :request-tar little-slave mir/ror '>' $(tests:get-tmp-dir)/archive.tar

tests:ensure tar -xlvf archive.tar
tests:assert-stdout-re "file_foo"
