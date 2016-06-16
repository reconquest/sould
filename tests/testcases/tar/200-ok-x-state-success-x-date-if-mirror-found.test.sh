#!/bin/bash

:sould-start little-slave --insecure

:git-repository upstream
:git-commit upstream file_foo

tests:ensure \
    :request-pull little-slave mir/ror $(tests:get-tmp-dir)/upstream

tests:ensure \
    :request-tar little-slave mir/ror '>' $(tests:get-tmp-dir)/archive.tar
tests:assert-stderr-re "200 OK"
tests:assert-stderr-re "X-State: success"
tests:assert-stderr-re "X-Date:"
