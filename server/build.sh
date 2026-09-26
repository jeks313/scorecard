#!/usr/bin/env bash
# Build the scorecard-server image with podman.
#
#   ./build.sh              # -> ghcr.io/jeks313/scorecard:latest
#   TAG=test ./build.sh     # tag :test instead
#   IMPORT=1 ./build.sh     # also load it straight into k3s containerd,
#                            # which is what lets the cluster run it without
#                            # the image ever leaving this box
set -euo pipefail
cd "$(dirname "$0")"

IMAGE="${IMAGE:-ghcr.io/jeks313/scorecard}"
TAG="${TAG:-latest}"
IMPORT="${IMPORT:-0}"

echo "Building ${IMAGE}:${TAG}"
podman build -t "${IMAGE}:${TAG}" -f Dockerfile .

if [[ $IMPORT == 1 ]]; then
    echo "==> importing into k3s containerd (k8s.io namespace)"
    podman save "${IMAGE}:${TAG}" | sudo k3s ctr --namespace k8s.io images import -
    sudo k3s ctr --namespace k8s.io images ls -q | grep -F "${IMAGE}:${TAG}"
fi

echo
echo "Built ${IMAGE}:${TAG}"
