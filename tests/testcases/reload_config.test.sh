#!/bin/bash

:sould-start jerk

@var sould_process :sould-process jerk
@var sould_stderr :sould-stderr jerk

tests:wait-file-changes $sould_stderr 0.1 3 kill -HUP $sould_process
tests:assert-re $sould_stderr "reloaded"

@var config :get-config jerk
tests:ensure rm $config

tests:wait-file-changes $sould_stderr 0.1 3 kill -HUP $sould_process

tests:assert-re "$sould_stderr" "can't reload config"
tests:assert-re "$sould_stderr" "no such file or directory"
