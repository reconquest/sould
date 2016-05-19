#!/bin/bash

:sould-start orphan

tests:ensure :request-status orphan hierarchical
tests:assert-no-diff stdout <<RESPONSE
status
├─ role: slave
└─ total: 0
RESPONSE
