:configure-master master 3000 pretty

:sould-start pretty --insecure
:sould-start master

:git-repository upstream
:git-commit     upstream foo

tests:ensure :request-pull pretty pool/x $(tests:get-tmp-dir)/upstream

@var storage_pretty :get-storage pretty
@var port_pretty :get-port pretty

tests:ensure rm -r $storage_pretty/pool/x/refs/

tests:ensure :request-status master hierarchical
tests:assert-no-diff stdout <<RESPONSE
status
├─ role: master
├─ total: 0
└─ upstream
   ├─ total: 1
   │
   ├─ success: 1 (100.00%)
   │
   ├─ error: 0 (0.00%)
   │
   └─ slaves
      └─ $(hostname):$port_pretty
         ├─ total: 1
         │
         ├─ error
         │  └─ can't get mirror pool/x
         │     └─ exec ["git" "config" "--get" "remote.origin.url"] error
         │        └─ exit status 1
         │           └─ output data is empty
         │
         └─ mirrors
            └─ pool/x
               └─ state: error
RESPONSE

tests:ensure :request-status master json
tests:assert-no-diff stdout <<RESPONSE
{
    "role": "master",
    "total": 0,
    "upstream": {
        "total": 1,
        "error": 0,
        "error_percent": 0,
        "success": 1,
        "success_percent": 100,
        "slaves": [
            {
                "address": "$(hostname):$port_pretty",
                "role": "slave",
                "total": 1,
                "error": "can't get mirror pool/x: exec [\"git\" \"config\" \"--get\" \"remote.origin.url\"] error (exit status 1) without output",
                "heararchical_error": "can't get mirror pool/x\n└─ exec [\"git\" \"config\" \"--get\" \"remote.origin.url\"] error\n   └─ exit status 1\n      └─ output data is empty",
                "mirrors": [
                    {
                        "name": "pool/x",
                        "state": "error"
                    }
                ]
            }
        ]
    }
}
RESPONSE

tests:ensure :request-status master toml
tests:assert-no-diff stdout <<RESPONSE
role = "master"
total = 0

[upstream]
    total = 1
    error = 0
    error_percent = 0.0
    success = 1
    success_percent = 100.0

    [[upstream.slaves]]
        address = "$(hostname):$port_pretty"
        error = "can't get mirror pool/x: exec [\"git\" \"config\" \"--get\" \"remote.origin.url\"] error (exit status 1) without output"
        heararchical_error = "can't get mirror pool/x\n└─ exec [\"git\" \"config\" \"--get\" \"remote.origin.url\"] error\n   └─ exit status 1\n      └─ output data is empty"
        role = "slave"
        total = 1

        [[upstream.slaves.mirrors]]
            name = "pool/x"
            state = "error"
RESPONSE
