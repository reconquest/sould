:configure-master master 3000 pretty sweety

:sould-start pretty --insecure
:sould-start master

:git-repository upstream
:git-commit     upstream foo

tests:ensure :request-pull pretty pool/x $(tests:get-tmp-dir)/upstream

@var storage_pretty :get-storage pretty
@var modify_date_pretty :git-modify-date $storage_pretty/pool/x
@var port_pretty :get-port pretty
@var port_sweety :get-port sweety

tests:ensure :request-status master hierarchical
tests:assert-no-diff stdout <<RESPONSE
status
├─ role: master
├─ total: 0
└─ upstream
   ├─ total: 2
   │
   ├─ success: 1 (50.00%)
   │
   ├─ error: 1 (50.00%)
   │
   └─ slaves
      ├─ $_hostname:$port_pretty
      │  ├─ role: slave
      │  │
      │  ├─ total: 1
      │  │
      │  └─ mirrors
      │     └─ pool/x
      │        ├─ state: success
      │        └─ modify date: $modify_date_pretty
      └─ $_hostname:$port_sweety
         ├─ role: slave
         │
         ├─ total: 0
         │
         └─ error
            └─ Get http://$_hostname:$port_sweety/x/status?format=json: dial tcp $_hostname_address:${port_sweety}: getsockopt: connection refused
RESPONSE

tests:ensure :request-status master json
tests:assert-no-diff stdout <<RESPONSE
{
    "role": "master",
    "total": 0,
    "upstream": {
        "total": 2,
        "error": 1,
        "error_percent": 50,
        "success": 1,
        "success_percent": 50,
        "slaves": [
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
            },
            {
                "address": "$_hostname:$port_sweety",
                "role": "slave",
                "total": 0,
                "error": "Get http://$_hostname:$port_sweety/x/status?format=json: dial tcp $_hostname_address:${port_sweety}: getsockopt: connection refused",
                "hierarchical_error": "Get http://$_hostname:$port_sweety/x/status?format=json: dial tcp $_hostname_address:${port_sweety}: getsockopt: connection refused"
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
    total = 2
    error = 1
    error_percent = 50.0
    success = 1
    success_percent = 50.0

    [[upstream.slaves]]
        address = "$_hostname:$port_pretty"
        role = "slave"
        total = 1

        [[upstream.slaves.mirrors]]
            modify_date = $modify_date_pretty
            name = "pool/x"
            state = "success"

    [[upstream.slaves]]
        address = "$_hostname:$port_sweety"
        error = "Get http://$_hostname:$port_sweety/x/status?format=json: dial tcp $_hostname_address:${port_sweety}: getsockopt: connection refused"
        hierarchical_error = "Get http://$_hostname:$port_sweety/x/status?format=json: dial tcp $_hostname_address:${port_sweety}: getsockopt: connection refused"
        role = "slave"
        total = 0
RESPONSE
