:sould-start orphan --insecure
@var storage :get-storage orphan

:git-repository upstream
:git-commit     upstream foo

tests:ensure :request-pull orphan pool/x $(tests:get-tmp-dir)/upstream
tests:ensure :request-pull orphan pool/y $(tests:get-tmp-dir)/upstream
tests:ensure :request-pull orphan pool/z $(tests:get-tmp-dir)/upstream

tests:ensure rm -r upstream
tests:eval :request-pull orphan pool/z $(tests:get-tmp-dir)/upstream

@var modify_date_x :git-modify-date $storage/pool/x
@var modify_date_y :git-modify-date $storage/pool/y
@var modify_date_z :git-modify-date $storage/pool/z

tests:ensure :request-status orphan hierarchical
tests:assert-no-diff stdout <<RESPONSE
status
├─ role: slave
│
├─ total: 3
│
└─ mirrors
   ├─ pool/x
   │  ├─ state: success
   │  └─ modify date: $modify_date_x
   ├─ pool/y
   │  ├─ state: success
   │  └─ modify date: $modify_date_y
   └─ pool/z
      ├─ state: error
      └─ modify date: $modify_date_z
RESPONSE

tests:ensure :request-status orphan json
tests:assert-no-diff stdout <<RESPONSE
{
    "role": "slave",
    "total": 3,
    "mirrors": [
        {
            "name": "pool/x",
            "state": "success",
            "modify_date": $modify_date_x
        },
        {
            "name": "pool/y",
            "state": "success",
            "modify_date": $modify_date_y
        },
        {
            "name": "pool/z",
            "state": "error",
            "modify_date": $modify_date_z
        }
    ]
}
RESPONSE

tests:ensure :request-status orphan toml
tests:assert-no-diff stdout <<RESPONSE
role = "slave"
total = 3

[[mirrors]]
    name = "pool/x"
    state = "success"
    modify_date = $modify_date_x

[[mirrors]]
    name = "pool/y"
    state = "success"
    modify_date = $modify_date_y

[[mirrors]]
    name = "pool/z"
    state = "error"
    modify_date = $modify_date_z
RESPONSE
