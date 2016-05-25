#!/bin/bash

:sould-start little-slave --insecure

:git-repository upstream
:git-commit upstream file_foo

tests:ensure \
    :request-pull little-slave mir/ror $(tests:get-tmp-dir)/upstream

tests:ensure \
    :request-tar little-slave mir/ror '>' $(tests:get-tmp-dir)/archive.tar
tests:assert-stderr-re "X-Date:"

@var header_x_date grep "X-Date:" $(tests:get-stderr-file)

tests:ensure mv upstream x

tests:ensure \
    :request-tar little-slave mir/ror ">" archive2.tar
tests:assert-stderr-re "$header_x_date"

:sould-stop little-slave
:sould-start little-slave --insecure

tests:ensure \
    :request-tar little-slave mir/ror ">" archive3.tar
tests:assert-stderr-re "$header_x_date"
