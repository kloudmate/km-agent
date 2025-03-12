#!/bin/bash

if [ -z "${KM_API_KEY}" ]; then
  echo "Environment variable KM_API_KEY is not set."
  exit 1
else
  echo "Environment variable KM_API_KEY is set to: $KM_API_KEY"
fi

VOLUME_MOUNTS=()
while [[ "$#" -gt 0 ]]; do
  case $1 in
    --volume|-v)
      VOLUME_MOUNTS+=("$2")
      shift 2
      ;;
    *)
      echo "Unknown : $1"
      exit 1
      ;;
  esac
done

VOLUME_ARGS=""
for VOLUME in "${VOLUME_MOUNTS[@]}"; do
  VOLUME_ARGS="$VOLUME_ARGS -v $VOLUME"
done

echo $VOLUME_ARGS

docker run -d\
  --name kmagent \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  $VOLUME_ARGS \
  -e KM_API_KEY=${KM_API_KEY} \
  --privileged \
  --pid=host \
  ghcr.io/kloudmate/km-agent:latest \
  /kmagent -mode=docker start