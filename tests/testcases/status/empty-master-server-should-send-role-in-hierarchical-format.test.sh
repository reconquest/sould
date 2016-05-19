#!/bin/bash

:configure-master grandma 3000

:sould-start grandma

tests:ensure :request-status grandma hierarchical
tests:assert-no-diff-blank stdout <<RESPONSE
status
├─ master
│  └─ total: 0
└─ upstream
   ├─ total: 0
   ├─ success: 0 (0.00%)
   └─ error: 0 (0.00%)
RESPONSE
