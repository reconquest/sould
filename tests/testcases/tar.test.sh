#!/bin/bash

:sould-start little-slave --insecure

:git-repository upstream
:git-commit upstream file_foo

mirror_name=mirror/of/upstream

tests:eval :request-tar little-slave $mirror_name
tests:assert-stderr-re "404 Not Found"

tests:eval :request-pull little-slave $mirror_name $(tests:get-tmp-dir)/upstream
tests:assert-stdout-re "200 OK"

tests:eval :request-tar little-slave $mirror_name '>' $(tests:get-tmp-dir)/archive.tar
tests:assert-stderr-re "200 OK"
tests:assert-stderr-re "X-State: success"
tests:assert-stderr-re "X-Date:"

@var header_x_date grep "X-Date:" $(tests:get-stderr-file)

tests:ensure tar -xlvf archive.tar
tests:assert-stdout-re "file_foo"

tests:ensure mv upstream backup_upstream

# sould should return tar archive with status 'success', because should
# not make pull request to repository.
tests:eval :request-tar little-slave $mirror_name ">" archive2.tar
tests:assert-stderr-re "200 OK"
tests:assert-stderr-re "X-State: success"
tests:assert-stderr-re "$header_x_date"

tests:ensure tar -xlvf archive2.tar
tests:assert-stdout-re 'file_foo'

# but if sould will be restarted, states table will be flushed, and state will
# be 'unknown', so sould should try to make pull, but pull will be failed
# (upstream corrupted earlier), and despite this, sould should return tar
# archive of last available version and show header 'X-Date' with last
# successfull update date.

:sould-stop little-slave
:sould-start little-slave --insecure

tests:eval :request-tar little-slave $mirror_name ">" archive3.tar
tests:assert-stderr-re "200 OK"
tests:assert-stderr-re "X-State: error"
tests:assert-stderr-re "$header_x_date"

tests:ensure tar -xlvf archive3.tar
tests:assert-stdout-re 'file_foo'
