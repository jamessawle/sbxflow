#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 3 ]; then
  echo "usage: $0 <version> <checksums-file> <output-file>" >&2
  exit 1
fi

version=$1
checksums_file=$2
output_file=$3

if [[ ! "${version}" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)([-+][0-9A-Za-z.-]+)?$ ]]; then
  echo "invalid release version: ${version}" >&2
  exit 1
fi

darwin_arm64="sbxflow_${version}_darwin_arm64.tar.gz"
darwin_amd64="sbxflow_${version}_darwin_amd64.tar.gz"

checksum_for() {
  local artifact=$1
  local checksum

  checksum=$(awk -v artifact="${artifact}" '$2 == artifact { print $1 }' "${checksums_file}")
  if [[ ! "${checksum}" =~ ^[0-9a-f]{64}$ ]]; then
    echo "missing or invalid checksum for ${artifact}" >&2
    exit 1
  fi

  printf '%s' "${checksum}"
}

arm64_checksum=$(checksum_for "${darwin_arm64}")
amd64_checksum=$(checksum_for "${darwin_amd64}")

mkdir -p "$(dirname "${output_file}")"

cat >"${output_file}" <<EOF
cask "sbxflow" do
  version "${version}"

  on_arm do
    sha256 "${arm64_checksum}"

    url "https://github.com/jamessawle/sbxflow/releases/download/v#{version}/sbxflow_#{version}_darwin_arm64.tar.gz"
  end
  on_intel do
    sha256 "${amd64_checksum}"

    url "https://github.com/jamessawle/sbxflow/releases/download/v#{version}/sbxflow_#{version}_darwin_amd64.tar.gz"
  end

  name "sbxflow"
  desc "Apply a repository's Docker Sandbox configuration and lifecycle"
  homepage "https://github.com/jamessawle/sbxflow"

  binary "sbxflow"
end
EOF
