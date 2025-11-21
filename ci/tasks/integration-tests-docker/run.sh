#!/usr/bin/env bash

set -euo pipefail

echo "Running Switchblade integration tests on Docker platform..."

export BUILDPACK_FILE=$(pwd)/built-buildpack/java-buildpack-*.zip
export GEM_HOME=$PWD/gems
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8
export LC_ALL=en_US.UTF-8

eval "$(rbenv init -)"

cd java-buildpack

echo "Installing Go dependencies..."
go mod download

echo "Running integration tests with Docker platform..."
go test -C src/integration -v \
  -platform=docker \
  -parallel=4 \
  -github-token="${GITHUB_TOKEN}" \
  -timeout=60m

echo "Docker integration tests completed successfully!"
