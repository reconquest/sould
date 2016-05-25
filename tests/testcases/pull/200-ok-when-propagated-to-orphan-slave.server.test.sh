#!/bin/bash

:sould-start orphan --insecure

:git-repository upstream
:git-commit     upstream foo

tests:ensure :request-pull orphan orphan/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re "200 OK"

@var storage :get-storage orphan
tests:test -d "$storage/orphan/fork"

:git-log "$storage/orphan/fork"
tests:assert-stdout-re "foo"
