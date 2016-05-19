#!/bin/bash

:configure-master grandma 3000

:sould-start grandma

tests:ensure :request-status grandma toml
tests:assert-no-diff-blank stdout <<RESPONSE
role = "master"

[master]
    total = 0

[upstream]
    total = 0
    error = 0
    error_percent = 0.0
    success = 0
    success_percent = 0.0
RESPONSE
