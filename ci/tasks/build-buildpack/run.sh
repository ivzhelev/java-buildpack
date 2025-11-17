#!/usr/bin/env bash

set -euo pipefail

cd java-buildpack

echo "Installing dependencies..."
bundle install

echo "Building online buildpack..."
OFFLINE=false bundle exec rake package

echo "Copying buildpack to output..."
VERSION=$(cat ../buildpack-version/version)
cp build/java-buildpack-*.zip ../built-buildpack/java-buildpack-v${VERSION}.zip

echo "Buildpack built successfully: java-buildpack-v${VERSION}.zip"
ls -lh ../built-buildpack/
