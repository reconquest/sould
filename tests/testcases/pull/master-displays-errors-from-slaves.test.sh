#!/bin/bash

:configure-master grandma 3000 grandson granddaughter

:sould-start grandson --insecure
:sould-start granddaughter --insecure
:sould-start grandma --insecure

:git-repository upstream
:git-commit     upstream foo

tests:ensure :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream

@var storage_grandson      :get-storage grandson
@var port_grandson         :get-port grandson
@var port_granddaughter    :get-port granddaughter

tests:ensure mv $storage_grandson backup_grandson
tests:ensure ln -s /dev/null $storage_grandson

tests:ensure mv upstream backup_upstream

tests:not tests:ensure :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re '< HTTP/1.1 503 Service Unavailable'

tests:ensure grep -A 5 -P "slave $_hostname:$port_grandson" response
tests:assert-stdout-re "500 Internal Server Error"
tests:assert-stdout-re "$storage_grandson/ma/fork: not a directory"

tests:ensure grep -A 1 -P "slave $_hostname:$port_granddaughter" response
tests:assert-stdout-re "500 Internal Server Error"
