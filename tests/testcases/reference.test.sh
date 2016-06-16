#!/bin/bash

:sould-start pretty-slave --insecure

:git-repository upstream
:git-commit upstream file_foo

@var reference :git-reference upstream

tests:ensure \
    :request-pull \
    pretty-slave super/upstream/fake $(tests:get-tmp-dir)/upstream

tests:ensure \
    :request-tar pretty-slave super/upstream/fake '>' archive.tar

tests:ensure tar -xlvf $(tests:get-tmp-dir)/archive.tar
tests:assert-stdout-re "file_foo"

tests:ensure :git-commit "upstream" "file_bar"

tests:ensure \
    :request-pull \
    pretty-slave super/upstream/fake $(tests:get-tmp-dir)/upstream

tests:ensure \
    :request-tar pretty-slave super/upstream/fake '>' archive.tar

tests:ensure tar -xlvf $(tests:get-tmp-dir)/archive.tar
tests:assert-stdout-re "file_foo"
tests:assert-stdout-re "file_bar"

tests:ensure \
    :request-tar \
    pretty-slave super/upstream/fake "$reference" '>' archive.tar

tests:ensure tar -xlvf $(tests:get-tmp-dir)/archive.tar
tests:assert-stdout-re "file_foo"
tests:not tests:assert-stdout-re "file_bar"
