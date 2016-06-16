#!/bin/bash

:configure-master grandma 3000 grandson granddaughter

:sould-start grandson --insecure
:sould-start granddaughter --insecure
:sould-start grandma --insecure

:git-repository upstream
:git-commit     upstream foo

tests:ensure \
    :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re "200 OK"

tests:ensure mv upstream backup_upstream

tests:not tests:ensure \
    :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re '< HTTP/1.1 503 Service Unavailable'
