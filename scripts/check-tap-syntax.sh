#!/usr/bin/env bash

# Run the Homebrew tap's own syntax checks against a generated formula.
#
# The tap's CI runs `brew test-bot --only-tap-syntax`, which is `brew style`,
# `brew readall --os=all --arch=all`, and `brew audit`. Those checks depend on
# the Homebrew version in use, so this runs them on a current Homebrew inside a
# throwaway container rather than against whatever the host happens to have
# installed. This is why the check is opt-in rather than part of `validate`: it
# needs Docker and network access and takes a few minutes.

set -euo pipefail

formula=${1:-dist/homebrew/Formula/sbxflow.rb}

if [ ! -f "${formula}" ]; then
  echo "no formula at ${formula}; run 'mise run test:release' first" >&2
  exit 1
fi

if ! docker version >/dev/null 2>&1; then
  echo "Docker is required to run the tap syntax checks" >&2
  exit 1
fi

staging=$(mktemp -d)
trap 'rm -rf "${staging}"' EXIT
mkdir -p "${staging}/homebrew-tap/Formula"
cp "${formula}" "${staging}/homebrew-tap/Formula/"

# COPYFILE_DISABLE stops macOS tar emitting AppleDouble `._` files, which
# Homebrew's Ruby linter reads as formulae with invalid byte sequences.
COPYFILE_DISABLE=1 tar -c -C "${staging}" homebrew-tap | docker run -i --rm \
  -e HOMEBREW_NO_SANDBOX_LINUX=1 \
  -e HOMEBREW_NO_REQUIRE_TAP_TRUST=1 \
  ubuntu:24.04 bash -euo pipefail -c '
    export DEBIAN_FRONTEND=noninteractive
    apt-get update -qq >/dev/null
    # build-essential is required: brew audit installs its RubyGems bundle, and
    # gems with native extensions cannot build without a compiler.
    apt-get install -y -qq git curl file procps ca-certificates sudo build-essential >/dev/null

    useradd -m -s /bin/bash brewuser
    echo "brewuser ALL=(ALL) NOPASSWD:ALL" >>/etc/sudoers

    mkdir -p /tmp/incoming && tar -x -C /tmp/incoming
    chown -R brewuser /tmp/incoming

    sudo -u brewuser -H bash -euo pipefail -c "
      NONINTERACTIVE=1 /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\" >/dev/null
      export PATH=/home/linuxbrew/.linuxbrew/bin:\$PATH
      export HOMEBREW_NO_AUTO_UPDATE=1
      export HOMEBREW_NO_ENV_HINTS=1

      echo \"Checking against \$(brew --version | head -1)\"

      tap=\"\$(brew --repository)/Library/Taps/jamessawle/homebrew-tap\"
      mkdir -p \"\$(dirname \"\$tap\")\"
      cp -r /tmp/incoming/homebrew-tap \"\$tap\"

      rc=0
      brew style jamessawle/tap || rc=1
      brew readall --aliases --os=all --arch=all jamessawle/tap || rc=1
      brew audit --except=installed --tap=jamessawle/tap || rc=1

      if [ \"\$rc\" -ne 0 ]; then
        echo \"Generated formula does not pass the tap syntax checks.\" >&2
      else
        echo \"Generated formula passes the tap syntax checks.\"
      fi
      exit \$rc
    "
  '
