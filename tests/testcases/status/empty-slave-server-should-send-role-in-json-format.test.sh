#!/bin/bash

:sould-start orphan

tests:ensure :request-status orphan json
tests:assert-no-diff stdout <<RESPONSE
{
    "role": "slave",
    "total": 0
}
RESPONSE
