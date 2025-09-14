#!/usr/bin/env bash
# Install Ambros reproducibly. By default installs @latest but forces direct proxy.
# Usage: ./scripts/install_ambros.sh [v3.1.2|latest]

set -euo pipefail
VERSION=${1:-latest}

echo "Installing ambros version: ${VERSION} (forced direct fetch)"
GOPROXY=direct go install github.com/gi4nks/ambros/v3@${VERSION}

echo "Installed: $(command -v ambros)"
ambros version || true
