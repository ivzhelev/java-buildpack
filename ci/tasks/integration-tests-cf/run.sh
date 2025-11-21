#!/usr/bin/env bash

set -euo pipefail

echo "Running Switchblade integration tests on Cloud Foundry platform..."

export BUILDPACK_FILE=$(pwd)/built-buildpack/java-buildpack-*.zip
export GEM_HOME=$PWD/gems
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8
export LC_ALL=en_US.UTF-8

# Initialize Ruby version manager (support both rbenv and mise)
if command -v rbenv &> /dev/null; then
  eval "$(rbenv init -)"
elif command -v mise &> /dev/null; then
  eval "$(mise activate bash)"
else
  echo "Warning: Neither rbenv nor mise found, using system Ruby"
fi

# Extract Cloud Foundry credentials from BBL state
echo "Extracting Cloud Foundry credentials from BBL state..."
cd bbl-state
eval "$(bbl print-env)"

# Get CF admin password from deployment vars
CF_ADMIN_PASSWORD=$(bosh int deployment-vars.yml --path /cf_admin_password)

cd ../java-buildpack

# Set up CF environment variables for Switchblade
export CF_API="https://api.${SYSTEM_DOMAIN}"
export CF_USERNAME="admin"
export CF_PASSWORD="${CF_ADMIN_PASSWORD}"
export CF_SKIP_SSL_VALIDATION="true"

echo "Installing Go dependencies..."
go mod download

echo "Running integration tests with Cloud Foundry platform..."
go test -C src/integration -v \
  -platform=cf \
  -stack=cflinuxfs4 \
  -parallel=4 \
  -github-token="${GITHUB_TOKEN}" \
  -timeout=90m

echo "Cloud Foundry integration tests completed successfully!"
