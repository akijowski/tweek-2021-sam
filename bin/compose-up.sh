#!/usr/bin/env bash

set -euo pipefail

function cleanup() {
    docker-compose down
}

trap cleanup EXIT

docker-compose up \
    --build \
    --attach-dependencies