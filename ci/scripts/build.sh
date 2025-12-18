#!/bin/bash -eux

# Build the application
pushd pull_request
  make build
  cp build/dis-redirect-api Dockerfile.concourse ../build
popd
