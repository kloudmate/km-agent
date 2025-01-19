#!/bin/bash

if [ -z "${KM_API_KEY}" ]; then
  echo "Environment variable KM_API_KEY is not set."
  exit 1
else
  echo "Environment variable KM_API_KEY is set to: $KM_API_KEY"
fi

docker run -d\
  --name kmagent \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  -v /proc:/hostfs/proc:ro \
  -v /sys:/hostfs/sys:ro \
  -v /:/hostfs:ro \
  --privileged \
  --pid=host \
  ghcr.io/kloudmate/km-agent:latest \
  /kmagent -m=docker -t=${KM_API_KEY} start