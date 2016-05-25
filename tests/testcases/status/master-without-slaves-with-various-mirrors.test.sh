:configure-master master 3000
:sould-start master --insecure
@var storage :get-storage master

:git-repository upstream
:git-commit     upstream foo

tests:ensure \
	:request-pull master pool/x $(tests:get-tmp-dir)/upstream
tests:ensure \
	:request-pull master pool/y $(tests:get-tmp-dir)/upstream
tests:ensure \
	:request-pull master pool/z $(tests:get-tmp-dir)/upstream

tests:ensure rm -r upstream
tests:eval :request-pull master pool/z $(tests:get-tmp-dir)/upstream

@var modify_date_x :git-modify-date $storage/pool/x
@var modify_date_y :git-modify-date $storage/pool/y
@var modify_date_z :git-modify-date $storage/pool/z

tests:ensure \
	:request-status master hierarchical
tests:assert-no-diff stdout <<RESPONSE
status
├─ role: master
│
├─ total: 3
│
├─ mirrors
│  ├─ pool/x
│  │  ├─ state: success
│  │  └─ modify date: $modify_date_x
│  ├─ pool/y
│  │  ├─ state: success
│  │  └─ modify date: $modify_date_y
│  └─ pool/z
│     ├─ state: error
│     └─ modify date: $modify_date_z
│
└─ upstream
   ├─ total: 0
   ├─ success: 0 (0.00%)
   └─ error: 0 (0.00%)
RESPONSE

tests:ensure \
	:request-status master json
tests:assert-no-diff stdout <<RESPONSE
{
    "role": "master",
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
    ],
    "upstream": {
        "total": 0,
        "error": 0,
        "error_percent": 0,
        "success": 0,
        "success_percent": 0
    }
}
RESPONSE

tests:ensure \
	:request-status master toml
tests:assert-no-diff stdout <<RESPONSE
role = "master"
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

[upstream]
    total = 0
    error = 0
    error_percent = 0.0
    success = 0
    success_percent = 0.0
RESPONSE
