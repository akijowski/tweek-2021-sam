#!/usr/bin/env bash

function describe-table() {
    readonly key=$1
    readonly secret=$2
    readonly table=$3

    AWS_ACCESS_KEY_ID=$key AWS_SECRET_ACCESS_KEY=$secret aws dynamodb describe-table \
      --table-name "$table" \
      --endpoint-url http://localhost:8000 \
      --region us-east-1 | jq
}

describe-table "defaultkey" "defaultsecret" "$1"