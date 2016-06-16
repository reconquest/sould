:sould-start bob

tests:not tests:ensure \
    :request-pull bob "no/matter1" "string-without-url-scheme-prefix"
tests:assert-re response '400 Bad Request'

tests:not tests:ensure \
    :request-pull bob "no/matter2" "string-https://-with-url-scheme"
tests:assert-re response '400 Bad Request'

# should be 500, because :clone will be failed, when got not working url
tests:not tests:ensure \
    :request-pull bob "no/matter3" "https://some-wrong-url"
tests:assert-re response '500 Internal Server Error'
