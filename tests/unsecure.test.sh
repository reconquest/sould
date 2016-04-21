#!/bin/bash

set -u

# secure mode

config_secure="`get_config_slave $(get_listen_addr 0) $(get_storage 0)`"
tests_ensure run_sould $config_secure false

tests_do request_pull 0 "no/matter1" "string-without-url-scheme-prefix"
tests_assert_stdout_re '400 Bad Request'

tests_do request_pull 0 "no/matter2" "string-https://-with-url-scheme"
tests_assert_stdout_re '400 Bad Request'

# should be 500, because git clone will be failed, when got not working url
tests_do request_pull 0 "no/matter3" "https://some-wrong-url"
tests_assert_stdout_re '500 Internal Server Error'

# insecure mode

config_insecure="`get_config_slave $(get_listen_addr 1) $(get_storage 1)`"
tests_ensure run_sould $config_insecure true

tests_do request_pull 1 "no/matter3" "string-without-url-scheme-prefix"
tests_assert_stdout_re '500 Internal Server Error'

tests_do request_pull 1 "no/matter4" "string-https://-with-url-scheme"
tests_assert_stdout_re '500 Internal Server Error'

tests_do request_pull 1 "no/matter5" "https://some-wrong-url"
tests_assert_stdout_re '500 Internal Server Error'
