#!/bin/bash

url_count=60

for ((i = 1; i <= url_count; i++)); do
  url="https://httpbin.org/get?val=${i}"

  req_count=$((1 + $RANDOM % 5))

  for ((j = 1; j <= req_count; j++)); do
    curl -XPOST http://localhost:8080/api/v1/urls -d "{\"url\": \"${url}\"}"
  done
done
