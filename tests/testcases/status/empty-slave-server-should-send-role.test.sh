#!/bin/bash

:sould-start orphan

tests:ensure :request-status orphan toml
tests:assert-stdout 'role = "slave"'
