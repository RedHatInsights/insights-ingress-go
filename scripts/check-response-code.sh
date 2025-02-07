#!/usr/bin/env sh

set -e

servername=$1
expected_response_code=$2

exit_code=1

response_code=$(curl --write-out %{http_code} -s -o /dev/null $servername)

if [ $expected_response_code = $response_code ]; then
  exit_code=0
fi

exit $exit_code
