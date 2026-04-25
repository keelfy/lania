#!/bin/bash
set -euo pipefail

MC_USER="minecraft"
MC_DIR="/opt/minecraft"
MRPACK_URL="$1"

if [ -z "$MRPACK_URL" ]; then
  echo "Usage: ./update.sh <mrpack_url>"
  exit 1
fi

echo ">>> Stopping server if running..."
sudo systemctl is-active --quiet minecraft && sudo systemctl stop minecraft || true

echo ">>> Backing up world..."
cp -r "$MC_DIR/world" "$MC_DIR/world_backup_$(date +%Y%m%d_%H%M%S)"

echo ">>> Updating modpack..."
sudo mrpack-install \
  --server-dir "$MC_DIR" \
  --server-file server.jar \
  "$MRPACK_URL"

sudo chown -R "$MC_USER":"$MC_USER" "$MC_DIR"

echo ">>> Starting server..."
sudo systemctl start "$MC_USER"

echo ">>> Done! Logs:"
sudo journalctl -u "$MC_USER" -f
