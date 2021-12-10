#!/usr/bin/env bash

function validate() {
    readonly okta_profile=$1
    readonly region=$2

    aws-okta exec "$okta_profile" -- sam validate --region "$region"
}

validate preprod_eng_customergrowth us-east-2