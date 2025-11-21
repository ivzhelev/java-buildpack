#!/usr/bin/env bash

set -euo pipefail

cd java-buildpack

echo "Installing dependencies..."
bundle install

echo "Building online buildpack for testing..."
bundle exec rake package PINNED=false

echo "Copying buildpack to output..."
cp build/java-buildpack-*.zip ../built-buildpack/java-buildpack-dev.zip

echo "Buildpack built successfully: java-buildpack-dev.zip"
ls -lh ../built-buildpack/
