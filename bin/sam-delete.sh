#!/usr/bin/env bash

function delete-stack() {
    readonly okta_profile=$1
    readonly region=$2
    readonly stackname=$3

    aws-okta exec "$okta_profile" -- sam delete --region "$region" --stack-name "$stackname"
}

delete-stack preprod_eng_customergrowth us-east-2 "$1"