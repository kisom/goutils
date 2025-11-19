#!/usr/bin/env bash

# Release Docker image for kisom/goutils using the Dockerfile in the repo root.
#
# Behavior:
# - Determines the git tag that points to HEAD. If no tag points to HEAD, aborts.
# - Builds the Docker image from the top-level Dockerfile.
# - Tags the image as kisom/goutils:<TAG> and kisom/goutils:latest.
# - Pushes both tags.
#
# Usage:
#   ./release-docker.sh

set -euo pipefail

err() { printf "Error: %s\n" "$*" >&2; }
info() { printf "==> %s\n" "$*"; }

# Ensure we're inside a git repository and operate from the repo root.
if ! REPO_ROOT=$(git rev-parse --show-toplevel 2>/dev/null); then
  err "This script must be run within a git repository."
  exit 1
fi
cd "$REPO_ROOT"

IMAGE_REPO="kisom/goutils"
DOCKERFILE_PATH="$REPO_ROOT/Dockerfile"

if [[ ! -f "$DOCKERFILE_PATH" ]]; then
  err "Dockerfile not found at repository root: $DOCKERFILE_PATH"
  err "Create a top-level Dockerfile or adjust this script before releasing."
  exit 1
fi

# Find tags that point to HEAD.
if ! TAGS=$(git tag --points-at HEAD); then
  err "Unable to query git tags."
  exit 1
fi

if [[ -z "$TAGS" ]]; then
  err "No git tag points at HEAD. Aborting release."
  exit 1
fi

# Use the first tag if multiple are present; warn the user.
# Avoid readarray for broader Bash compatibility (e.g., macOS Bash 3.2).
TAG_ARRAY=($TAGS)
TAG="${TAG_ARRAY[0]}"

if (( ${#TAG_ARRAY[@]} > 1 )); then
  info "Multiple tags point at HEAD: ${TAG_ARRAY[*]}"
  info "Using first tag: $TAG"
fi

info "Releasing Docker image for tag: $TAG"

IMAGE_TAGGED="$IMAGE_REPO:$TAG"
IMAGE_LATEST="$IMAGE_REPO:latest"

info "Building image from $DOCKERFILE_PATH"
docker build -f "$DOCKERFILE_PATH" -t "$IMAGE_TAGGED" -t "$IMAGE_LATEST" "$REPO_ROOT"

info "Pushing $IMAGE_TAGGED"
docker push "$IMAGE_TAGGED"

info "Pushing $IMAGE_LATEST"
docker push "$IMAGE_LATEST"

info "Release complete."
