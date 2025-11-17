#!/usr/bin/env bash

set -euo pipefail

echo "Extracting Cloud Foundry credentials from BBL state..."
cd bbl-state

# Source BBL environment variables
eval "$(bbl print-env)"

# Get CF admin password from deployment vars
CF_ADMIN_PASSWORD=$(bosh int deployment-vars.yml --path /cf_admin_password)

cd ../java-buildpack-system-test

echo "Logging into Cloud Foundry..."
cf api "https://api.${SYSTEM_DOMAIN}" --skip-ssl-validation
cf auth admin "${CF_ADMIN_PASSWORD}"

echo "Creating test organization and space..."
cf create-org java-buildpack-test || true
cf target -o java-buildpack-test
cf create-space test || true
cf target -s test

echo "Uploading buildpack..."
BUILDPACK_PATH=$(ls ../built-buildpack/java-buildpack-*.zip)
cf create-buildpack java_buildpack_test "${BUILDPACK_PATH}" 1 --enable || \
  cf update-buildpack java_buildpack_test -p "${BUILDPACK_PATH}" --enable

echo "Running system tests..."
./mvnw clean test -Dtest.java.buildpack=java_buildpack_test

echo "Cleaning up test resources..."
cf delete-buildpack java_buildpack_test -f || true
cf delete-space test -f || true
cf delete-org java-buildpack-test -f || true

echo "System tests completed successfully!"
