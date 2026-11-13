#!/bin/bash -eux

# Configure Git to use the provided GitHub credentials for accessing internal repositories
set +x
git config --global url."https://${GITHUB_USERNAME}:${GITHUB_TOKEN}@github.com/${GITHUB_REPO_PATH}".insteadOf "https://github.com/${GITHUB_REPO_PATH}"
set -x

# Build the application
pushd pull_request
  make build-go
  cp build/dis-redirect-api Dockerfile.concourse ../build
popd
