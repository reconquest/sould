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

@var storage_grandson :get-storage grandson

tests:ensure mv $storage_grandson backup_grandson
tests:ensure ln -s /dev/null $storage_grandson

tests:not tests:ensure \
	:request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re '< HTTP/1.1 502 Bad Gateway'
tests:assert-stdout-re "^[^<].*500 Internal Server Error"
tests:assert-stdout-re "$storage_grandson/ma/fork: not a directory"
