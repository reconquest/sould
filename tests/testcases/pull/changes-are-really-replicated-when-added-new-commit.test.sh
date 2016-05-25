#!/bin/bash

:configure-master grandma 3000 grandson granddaughter

:sould-start grandson --insecure
:sould-start granddaughter --insecure
:sould-start grandma --insecure

:git-repository upstream
:git-commit     upstream foo

tests:ensure :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream

@var storage_grandson      :get-storage grandson
@var storage_granddaughter :get-storage granddaughter
@var storage_grandma       :get-storage grandma

tests:test -d "$storage_grandson/ma/fork"
:git-log "$storage_grandson/ma/fork"
tests:assert-stdout-re "foo"

tests:test -d "$storage_granddaughter/ma/fork"
:git-log "$storage_granddaughter/ma/fork"
tests:assert-stdout-re "foo"

tests:test -d "$storage_grandma/ma/fork"
:git-log "$storage_grandma/ma/fork"
tests:assert-stdout-re "foo"

:git-commit upstream bar

tests:ensure :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re '200 OK'

tests:test -d "$storage_grandson/ma/fork"
:git-log "$storage_grandson/ma/fork"
tests:assert-stdout-re "bar"

tests:test -d "$storage_granddaughter/ma/fork"
:git-log "$storage_granddaughter/ma/fork"
tests:assert-stdout-re "bar"

tests:test -d "$storage_grandma/ma/fork"
:git-log "$storage_grandma/ma/fork"
tests:assert-stdout-re "bar"
