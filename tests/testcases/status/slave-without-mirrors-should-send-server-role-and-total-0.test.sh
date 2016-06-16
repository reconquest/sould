#!/bin/bash

:sould-start orphan

tests:ensure \
    :request-status orphan hierarchical
tests:assert-no-diff stdout <<RESPONSE
status
├─ role: slave
└─ total: 0
RESPONSE

tests:ensure \
    :request-status orphan json
tests:assert-no-diff stdout <<RESPONSE
{
    "role": "slave",
    "total": 0
}
RESPONSE

tests:ensure \
    :request-status orphan toml
tests:assert-no-diff stdout <<RESPONSE
role = "slave"
total = 0
RESPONSE
