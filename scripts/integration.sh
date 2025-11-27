#!/usr/bin/env bash

set -euo pipefail

ROOTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly ROOTDIR

function usage() {
  cat <<-USAGE
integration.sh --github-token <token> [OPTIONS]

Runs the integration tests.

OPTIONS
  --help                  -h  prints the command usage
  --github-token <token>      GitHub token to use when making API requests
  --platform <cf|docker>      Switchblade platform to execute the tests against (default: cf)
  --cached <true|false>       Run cached/offline tests (default: false)
  --parallel <true|false>     Run tests in parallel (default: false)
  --stack <stack>             Stack to use for tests (default: cflinuxfs4)

EXAMPLES
  # Serial mode
  ./scripts/integration.sh --platform docker

  # Parallel mode (uses GOMAXPROCS=2)
  ./scripts/integration.sh --platform docker --parallel true
USAGE
}

function main() {
  local src stack platform token cached parallel
  src="${ROOTDIR}/src/integration"
  stack="${CF_STACK:-cflinuxfs4}"
  platform="cf"
  cached="false"
  parallel="false"
  token="${GITHUB_TOKEN:-}"

  while [[ "${#}" != 0 ]]; do
    case "${1}" in
      --platform)
        platform="${2}"
        shift 2
        ;;

      --github-token)
        token="${2}"
        shift 2
        ;;

      --cached)
        cached="${2}"
        shift 2
        ;;

      --parallel)
        parallel="${2}"
        shift 2
        ;;

      --stack)
        stack="${2}"
        shift 2
        ;;

      --help|-h)
        shift 1
        usage
        exit 0
        ;;

      "")
        # skip if the argument is empty
        shift 1
        ;;

      *)
        echo "ERROR: unknown argument \"${1}\""
        usage
        exit 1
        ;;
    esac
  done

  if [[ -z "${BUILDPACK_FILE:-}" ]]; then
    echo "ERROR: BUILDPACK_FILE environment variable is required"
    exit 1
  fi

  if [[ ! -f "${BUILDPACK_FILE}" ]]; then
    echo "ERROR: Buildpack file not found: ${BUILDPACK_FILE}"
    exit 1
  fi

  echo "=== Java Buildpack Integration Tests ==="
  echo "Platform:      ${platform}"
  echo "Stack:         ${stack}"
  echo "Cached:        ${cached}"
  echo "Parallel:      ${parallel}"
  echo "Buildpack:     ${BUILDPACK_FILE}"
  echo ""

  specs::run "${cached}" "${parallel}" "${stack}" "${platform}" "${token}"
}

function specs::run() {
  local cached parallel stack platform token
  cached="${1}"
  parallel="${2}"
  stack="${3}"
  platform="${4}"
  token="${5}"

  local nodes cached_flag serial_flag platform_flag stack_flag token_flag
  cached_flag="--cached=${cached}"
  serial_flag="--serial=true"
  platform_flag="--platform=${platform}"
  stack_flag="--stack=${stack}"
  token_flag="--github-token=${token}"
  nodes=1

  if [[ "${parallel}" == "true" ]]; then
    nodes=3
    serial_flag=""
  fi

  cd "${ROOTDIR}"
  go mod download

  CF_STACK="${stack}" \
  BUILDPACK_FILE="${BUILDPACK_FILE}" \
  GOMAXPROCS="${GOMAXPROCS:-"${nodes}"}" \
    go test \
      -count=1 \
      -timeout=0 \
      -mod vendor \
      -v \
        "${ROOTDIR}/src/integration" \
         ${cached_flag} \
         ${platform_flag} \
         ${token_flag} \
         ${stack_flag} \
         ${serial_flag}
}

main "${@:-}"
