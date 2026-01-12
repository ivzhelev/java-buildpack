#!/usr/bin/env bash

set -e
set -u
set -o pipefail

ROOTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
readonly ROOTDIR

# shellcheck source=SCRIPTDIR/.util/tools.sh
source "${ROOTDIR}/scripts/.util/tools.sh"

function main() {
  util::tools::jq::install --directory "${ROOTDIR}/.bin"

  IFS=" " read -r -a oses <<< "$(jq -r -S '.oses[]' "${ROOTDIR}/config.json" | xargs)"
  IFS=" " read -r -a arches <<< "$(jq -r -S '.arches[]' "${ROOTDIR}/config.json" | xargs)"

  mapfile -t binaries < <(find "${ROOTDIR}/src" -mindepth 2 -name cli -type d)

  for os in "${oses[@]}"; do
    for arch in "${arches[@]}"; do
      for path in "${binaries[@]}"; do
        local name output
        name="$(basename "$(dirname "${path}")")"
        output="${ROOTDIR}/bin/${name}"

        if [[ "${os}" == "windows" ]]; then
          output="${output}.exe"
        fi

        CGO_ENABLED=0 \
        GOOS="${os}" \
        GOARCH="${arch}" \
          go build \
            -mod vendor \
            -ldflags="-s -w" \
            -o "${output}" \
              "${path}"
      done
    done
  done
}

main "${@:-}"
