#!/usr/bin/env bash
set -euo pipefail

registry="${REGISTRY:-}"
namespace="${IMAGE_NAMESPACE:-chdxk}"
tag="${IMAGE_TAG:-latest}"
api_image="${API_IMAGE:-${registry:+${registry}/}${namespace}/api:${tag}}"
taskworker_image="${TASKWORKER_IMAGE:-${registry:+${registry}/}${namespace}/taskworker:${tag}}"

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(dirname -- "${script_dir}")"
docker build -f "${script_dir}/api.Dockerfile" -t "${api_image}" "${repo_root}"
docker build -f "${script_dir}/taskworker.Dockerfile" -t "${taskworker_image}" "${script_dir}"

if [[ "${PUSH_IMAGES:-false}" == "true" ]]; then
  docker push "${api_image}"
  docker push "${taskworker_image}"
fi
