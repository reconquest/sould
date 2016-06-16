#!/bin/bash

:sould-start little-slave --insecure

:git-repository upstream
:git-commit upstream file_foo

tests:ensure \
    :request-pull little-slave mir/ror $(tests:get-tmp-dir)/upstream

tests:ensure mv upstream xx

tests:not tests:ensure \
    :request-pull little-slave mir/ror $(tests:get-tmp-dir)/upstream

tests:ensure \
    :request-tar little-slave mir/ror ">" archive3.tar
tests:assert-stderr-re "X-State: error"
