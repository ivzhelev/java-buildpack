#!/usr/bin/env bash

set -euo pipefail

cd java-buildpack

echo "Installing dependencies..."
bundle install

echo "Building online buildpack for testing..."
OFFLINE=false bundle exec rake package

echo "Copying buildpack to output..."
cp build/java-buildpack-*.zip ../built-buildpack/java-buildpack-dev.zip

echo "Buildpack built successfully for tests"
ls -lh ../built-buildpack/
