#!/usr/bin/env bash

function create-table() {
    readonly key=$1
    readonly secret=$2
    readonly table=$3

    AWS_ACCESS_KEY_ID=$key AWS_SECRET_ACCESS_KEY=$secret aws dynamodb create-table \
      --table-name "$table" \
      --attribute-definitions AttributeName=owner,AttributeType=S AttributeName=title,AttributeType=S \
      --key-schema AttributeName=owner,KeyType=HASH AttributeName=title,KeyType=RANGE \
      --billing-mode PAY_PER_REQUEST \
      --endpoint-url http://localhost:8000 \
      --region us-east-1 | jq
}

create-table "defaultkey" "defaultsecret" "$1"