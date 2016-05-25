#!/bin/bash

:configure-master master 3000 pretty-slave

:sould-start pretty-slave --insecure
:sould-start master --insecure

:git-repository upstream

:git-commit upstream commit_A
:git-commit upstream commit_B

:git-tag    upstream tag-before-c
:git-commit upstream commit_C
:git-tag    upstream tag-after-c
:git-reset  upstream tag-before-c
:git-tag    upstream tag-before-c -d

tests:ensure \
    :request-pull-with-spoof \
    master \
    mirror/for/upstream \
    $(tests:get-tmp-dir)/upstream \
    master \
    tag-after-c

@var storage_slave :get-storage pretty-slave

:git-log $storage_slave/mirror/for/upstream
tests:assert-stdout-re commit_A
tests:assert-stdout-re commit_B
tests:assert-stdout-re commit_C

@var storage_master :get-storage master

:git-log $storage_master/mirror/for/upstream
tests:assert-stdout-re commit_A
tests:assert-stdout-re commit_B
tests:assert-stdout-re commit_C

