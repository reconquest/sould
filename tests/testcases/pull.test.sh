#!/bin/bash


:configure-master grandma 3000 grandson granddaughter

:sould-start grandson --insecure
:sould-start granddaughter --insecure
:sould-start grandma --insecure

:git-repository upstream
:git-commit     upstream foo


tests:ensure :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re "200 OK"

@var storage_grandson      :get-storage grandson
@var storage_granddaughter :get-storage granddaughter
@var storage_grandma       :get-storage grandma
@var port_grandson         :get-port grandson
@var port_granddaughter    :get-port granddaughter

# check for successfully propagating request to slaves, check commit in all
# storages
storages=("$storage_grandma" "$storage_grandson" "$storage_granddaughter")
for storage in $storages; do
    tests:test -d "$storage/ma/fork"

    :git-log "$storage/ma/fork"
    tests:assert-stdout-re "foo"
done

:git-commit upstream bar

tests:ensure :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re '200 OK'

for storage in $storages; do
    tests:test -d "$storage/ma/fork"

    :git-log "$storage/ma/fork"
    tests:assert-stdout-re "foo"
    tests:assert-stdout-re "bar"
done

# corrupting 'pa' slave storage
tests:ensure mv $storage_grandson backup_grandson
tests:ensure ln -s /dev/null $storage_grandson

tests:eval :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re '< HTTP/1.1 502 Bad Gateway'
tests:assert-stdout-re "^[^<].*500 Internal Server Error"
tests:assert-stdout-re "$storage_grandson/ma/fork: not a directory"

# corrupting upstream
tests:ensure mv upstream backup_upstream

tests:eval :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re '< HTTP/1.1 503 Service Unavailable'

tests:ensure grep -A 5 -P "slave $(hostname):$port_grandson" response
tests:assert-stdout-re "500 Internal Server Error"
tests:assert-stdout-re "$storage_grandson/ma/fork: not a directory"

tests:ensure grep -A 1 -P "slave $(hostname):$port_granddaughter" response
tests:assert-stdout-re "500 Internal Server Error"

#restoring sould upstream and storage_pa directory
tests:ensure mv backup_upstream upstream

tests:ensure rm -f $storage_grandson
tests:ensure mv backup_grandson $storage_grandson

tests:ensure :request-pull grandma ma/fork $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re '200 OK'
