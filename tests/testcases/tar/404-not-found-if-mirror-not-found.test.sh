#!/bin/bash

:sould-start little-slave --insecure

:git-repository upstream
:git-commit upstream file_foo

tests:not tests:ensure \
	:request-tar little-slave mir/ror
tests:assert-stderr-re "404 Not Found"
