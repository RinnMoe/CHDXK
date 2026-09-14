#!/usr/bin/env bash
set -euo pipefail

registry="${REGISTRY:-}"
namespace="${IMAGE_NAMESPACE:-chdxk}"
tag="${IMAGE_TAG:-latest}"
api_image="${API_IMAGE:-${registry:+${registry}/}${namespace}/api:${tag}}"
taskworker_image="${TASKWORKER_IMAGE:-${registry:+${registry}/}${namespace}/taskworker:${tag}}"

docker build -f api.Dockerfile -t "${api_image}" .
docker build -f taskworker.Dockerfile -t "${taskworker_image}" .

if [[ "${PUSH_IMAGES:-false}" == "true" ]]; then
  docker push "${api_image}"
  docker push "${taskworker_image}"
fi
