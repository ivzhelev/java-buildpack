#!/usr/bin/env bash

set -euo pipefail

# Script to run integration tests using Switchblade framework
# Supports both Cloud Foundry and Docker platforms

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
SRC_DIR="${ROOT_DIR}/src/integration"

# Default configuration
PLATFORM="${PLATFORM:-cf}"
STACK="${STACK:-cflinuxfs4}"
CACHED="${CACHED:-false}"
GITHUB_TOKEN="${GITHUB_TOKEN:-}"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

function print_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Run integration tests for the Java buildpack using Switchblade"
    echo ""
    echo "Options:"
    echo "  -p, --platform PLATFORM   Platform to test against (cf or docker, default: cf)"
    echo "  -s, --stack STACK         Stack to use (default: cflinuxfs4)"
    echo "  -c, --cached              Run cached/offline tests"
    echo "  -t, --github-token TOKEN  GitHub API token for rate limiting"
    echo "  -h, --help                Show this help message"
    echo ""
    echo "Environment Variables:"
    echo "  BUILDPACK_FILE            Path to buildpack zip file (required)"
    echo "  PLATFORM                  Platform to test (cf or docker)"
    echo "  STACK                     Stack to use for tests"
    echo "  CACHED                    Run cached tests (true/false)"
    echo "  GITHUB_TOKEN              GitHub API token"
    echo ""
    echo "Examples:"
    echo "  # Test on Cloud Foundry with cflinuxfs4"
    echo "  BUILDPACK_FILE=/tmp/buildpack.zip $0"
    echo ""
    echo "  # Test on Docker"
    echo "  BUILDPACK_FILE=/tmp/buildpack.zip $0 --platform docker"
    echo ""
    echo "  # Run cached/offline tests"
    echo "  BUILDPACK_FILE=/tmp/buildpack.zip $0 --cached"
}

# Parse command-line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        -s|--stack)
            STACK="$2"
            shift 2
            ;;
        -c|--cached)
            CACHED="true"
            shift
            ;;
        -t|--github-token)
            GITHUB_TOKEN="$2"
            shift 2
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            print_usage
            exit 1
            ;;
    esac
done

# Validate required environment variables
if [[ -z "${BUILDPACK_FILE:-}" ]]; then
    echo -e "${RED}ERROR: BUILDPACK_FILE environment variable is required${NC}"
    echo ""
    print_usage
    exit 1
fi

if [[ ! -f "${BUILDPACK_FILE}" ]]; then
    echo -e "${RED}ERROR: Buildpack file not found: ${BUILDPACK_FILE}${NC}"
    exit 1
fi

# Print configuration
echo -e "${GREEN}=== Java Buildpack Integration Tests ===${NC}"
echo "Platform:      ${PLATFORM}"
echo "Stack:         ${STACK}"
echo "Cached Tests:  ${CACHED}"
echo "Buildpack:     ${BUILDPACK_FILE}"
echo ""

# Check dependencies
if ! command -v go &> /dev/null; then
    echo -e "${RED}ERROR: Go is not installed${NC}"
    exit 1
fi

echo -e "${YELLOW}Installing Go dependencies...${NC}"
cd "${ROOT_DIR}"
go mod download

# Build test flags
TEST_FLAGS="-v"
TEST_FLAGS="${TEST_FLAGS} -platform=${PLATFORM}"
TEST_FLAGS="${TEST_FLAGS} -stack=${STACK}"

if [[ "${CACHED}" == "true" ]]; then
    TEST_FLAGS="${TEST_FLAGS} -cached"
fi

if [[ -n "${GITHUB_TOKEN}" ]]; then
    TEST_FLAGS="${TEST_FLAGS} -github-token=${GITHUB_TOKEN}"
fi

# Run tests
echo -e "${YELLOW}Running integration tests...${NC}"
echo ""

cd "${SRC_DIR}"

if go test ${TEST_FLAGS} -timeout 30m ./...; then
    echo ""
    echo -e "${GREEN}✓ All integration tests passed!${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}✗ Integration tests failed${NC}"
    exit 1
fi
