#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
tmp_dir=$(mktemp -d "${TMPDIR:-/tmp}/archivist-adr-tools-compare.XXXXXX")

cleanup() {
  if [[ ${KEEP_ADR_TOOLS_COMPARE_TMP:-} != 1 ]]; then
    rm -rf "$tmp_dir"
  else
    echo "Kept temporary directory: $tmp_dir" >&2
  fi
}
trap cleanup EXIT

if [[ -z ${ADR_TOOLS_DIR:-} && -z ${ADR_TOOLS_CMD:-} ]]; then
  git clone --depth 1 https://github.com/npryce/adr-tools.git "$tmp_dir/adr-tools"
  export ADR_TOOLS_DIR="$tmp_dir/adr-tools"
fi

cd "$repo_root"
go test -count=1 ./internal/compat -run TestADRToolsCompatibility "$@"
