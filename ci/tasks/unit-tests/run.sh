#!/usr/bin/env bash

set -euo pipefail

cd java-buildpack

echo "Installing dependencies..."
bundle install

echo "Running unit tests..."
bundle exec rake
