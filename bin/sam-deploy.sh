#!/usr/bin/env bash

function deploy() {
    readonly okta_profile=$1
    readonly region=$2
    readonly stackname=$3

    aws-okta exec "$okta_profile" -- sam deploy \
      --guided \
      --region "$region" \
      --stack-name "$stackname" \
      --capabilities CAPABILITY_IAM
}

deploy preprod_eng_customergrowth us-east-2 "$1"