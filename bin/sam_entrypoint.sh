#!/bin/bash
set -o errexit

BASEDIR="$1"
echo -e "CWD: $BASEDIR"
/usr/local/bin/sam local start-lambda \
  --debug \
  --host 0.0.0.0 \
  --docker-network aws-sam-api \
  --docker-volume-basedir "${BASEDIR}" \
  --container-host-interface 0.0.0.0 \
  --container-host host.docker.internal \
  --env-vars local/env-acceptance.json