#!/bin/bash

:sould-start sweety-slave --insecure

:git-repository upstream
:git-commit     upstream foo

tests:ensure :request-pull sweety-slave mirror $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re "200 OK"

:git-server-start sweety-slave

:git-clone sweety-slave mirror attack_of_the_clones

:git-log attack_of_the_clones
tests:assert-stdout-re 'foo'

tests:ensure rm -r attack_of_the_clones

tests:ensure mv upstream backup_upstream

:git-clone sweety-slave mirror attack_of_the_clones

:git-log attack_of_the_clones
tests:assert-stdout-re 'foo'

tests:ensure rm -r attack_of_the_clones

tests:not tests:ensure :request-pull sweety-slave mirror $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re '500 Internal Server Error'

:git-clone-fail sweety-slave mirror attack_of_the_clones

tests:ensure mv $(tests:get-tmp-dir)/backup_upstream $(tests:get-tmp-dir)/upstream

:git-clone sweety-slave mirror attack_of_the_clones

tests:tmp_cd git-cloned

:git-log attack_of_the_clones
tests:assert-stdout-re 'foo'
