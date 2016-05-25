:configure-master master 3000 pretty sweety

:sould-start pretty --insecure
:sould-start sweety --insecure
:sould-start master

:git-repository upstream
:git-commit     upstream foo

tests:ensure \
    :request-pull pretty pool/x $(tests:get-tmp-dir)/upstream

@var storage_pretty :get-storage pretty
@var modify_date_pretty :git-modify-date $storage_pretty/pool/x
@var port_pretty :get-port pretty
@var port_sweety :get-port sweety

tests:ensure \
    :request-status master hierarchical
tests:assert-no-diff stdout <<RESPONSE
status
├─ role: master
├─ total: 0
└─ upstream
   ├─ total: 2
   │
   ├─ success: 2 (100.00%)
   │
   ├─ error: 0 (0.00%)
   │
   └─ slaves
      ├─ $_hostname:$port_sweety
      │  ├─ role: slave
      │  └─ total: 0
      └─ $_hostname:$port_pretty
         ├─ role: slave
         │
         ├─ total: 1
         │
         └─ mirrors
            └─ pool/x
               ├─ state: success
               └─ modify date: $modify_date_pretty
RESPONSE

tests:ensure \
    :request-status master json
tests:assert-no-diff stdout <<RESPONSE
{
    "role": "master",
    "total": 0,
    "upstream": {
        "total": 2,
        "error": 0,
        "error_percent": 0,
        "success": 2,
        "success_percent": 100,
        "slaves": [
            {
                "address": "$_hostname:$port_sweety",
                "role": "slave",
                "total": 0
            },
            {
                "address": "$_hostname:$port_pretty",
                "role": "slave",
                "total": 1,
                "mirrors": [
                    {
                        "name": "pool/x",
                        "state": "success",
                        "modify_date": $modify_date_pretty
                    }
                ]
            }
        ]
    }
}
RESPONSE

tests:ensure \
    :request-status master toml
tests:assert-no-diff stdout <<RESPONSE
role = "master"
total = 0

[upstream]
    total = 2
    error = 0
    error_percent = 0.0
    success = 2
    success_percent = 100.0

    [[upstream.slaves]]
        address = "$_hostname:$port_sweety"
        role = "slave"
        total = 0

    [[upstream.slaves]]
        address = "$_hostname:$port_pretty"
        role = "slave"
        total = 1

        [[upstream.slaves.mirrors]]
            modify_date = $modify_date_pretty
            name = "pool/x"
            state = "success"
RESPONSE
