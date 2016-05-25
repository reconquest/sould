:sould-start alice --insecure

tests:not tests:ensure \
	:request-pull alice "no/matter3" "string-without-url-scheme-prefix"
tests:assert-re response '500 Internal Server Error'

tests:not tests:ensure \
	:request-pull alice "no/matter4" "string-https://-with-url-scheme"
tests:assert-re response '500 Internal Server Error'

tests:not tests:ensure \
	:request-pull alice "no/matter5" "https://some-wrong-url"
tests:assert-re response '500 Internal Server Error'

tests:not tests:ensure \
	:request-pull alice "no/matter6" "/not/exist/path/"
tests:assert-re response '500 Internal Server Error'
