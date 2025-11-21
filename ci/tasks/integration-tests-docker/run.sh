#!/usr/bin/env bash

set -euo pipefail

echo "Running Switchblade integration tests on Docker platform..."

# Start Docker daemon (required for privileged containers in Concourse)
start_docker() {
  echo "Starting Docker daemon..."
  
  if docker info >/dev/null 2>&1; then
    echo "Docker is already running"
    return 0
  fi

  # Check for cgroup v2
  if [ ! -e /sys/fs/cgroup/cgroup.controllers ]; then
    echo "ERROR: cgroup v1 detected but no longer supported"
    echo "This infrastructure requires cgroup v2 (Ubuntu 22.04+/Kubernetes 1.25+)"
    exit 1
  fi

  mkdir -p /var/log /var/run /scratch/docker

  # Remount /proc/sys as read-write if needed
  if grep '/proc/sys\s\+\w\+\s\+ro,' /proc/mounts >/dev/null 2>&1; then
    mount -o remount,rw /proc/sys
  fi

  # Start dockerd in background (use system-provided version)
  dockerd --data-root /scratch/docker --mtu 1200 >/tmp/docker.log 2>&1 &
  echo $! > /tmp/docker.pid

  # Wait for Docker to be ready
  echo "Waiting for Docker daemon to be ready..."
  for i in {1..30}; do
    if docker info >/dev/null 2>&1; then
      echo "Docker daemon started successfully"
      return 0
    fi
    sleep 1
  done

  echo "ERROR: Docker daemon failed to start"
  cat /tmp/docker.log
  exit 1
}

stop_docker() {
  echo "Stopping Docker daemon..."
  if [ -f /tmp/docker.pid ]; then
    kill "$(cat /tmp/docker.pid)" 2>/dev/null || true
  fi
}

# Start Docker and ensure cleanup on exit
start_docker
trap stop_docker EXIT

# Resolve wildcard to actual buildpack filename with absolute path
export BUILDPACK_FILE=$(ls $(pwd)/built-buildpack/java-buildpack-*.zip | head -1)
export GEM_HOME=$PWD/gems
export LANG=en_US.UTF-8
export LANGUAGE=en_US.UTF-8
export LC_ALL=en_US.UTF-8

# Set Docker API version to match daemon requirements (API 1.44+)
# The Docker SDK v24 defaults to API 1.43, but can negotiate higher versions
export DOCKER_API_VERSION=1.44

# Initialize Ruby version manager (support both rbenv and mise)
if command -v rbenv &> /dev/null; then
  eval "$(rbenv init -)"
elif command -v mise &> /dev/null; then
  eval "$(mise activate bash)"
else
  echo "Warning: Neither rbenv nor mise found, using system Ruby"
fi

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
