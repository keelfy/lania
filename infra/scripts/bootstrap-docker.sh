#!/bin/bash
set -euo pipefail

DEPLOY_PATH="$1"

if ! command -v docker >/dev/null 2>&1; then
  echo ">>> Docker not found, installing..."
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
else
  echo ">>> Docker already installed, skipping."
fi

mkdir -p "$DEPLOY_PATH/compose"
