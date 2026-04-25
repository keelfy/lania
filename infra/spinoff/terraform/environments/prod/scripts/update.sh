#!/bin/bash
set -euo pipefail

MC_DIR="/opt/minecraft"
MRPACK_URL="$1"

if [ -z "$MRPACK_URL" ]; then
  echo "Usage: ./update.sh <mrpack_url>"
  exit 1
fi

echo ">>> Stopping server..."
systemctl stop minecraft

echo ">>> Backing up world..."
cp -r "$MC_DIR/world" "$MC_DIR/world_backup_$(date +%Y%m%d_%H%M%S)"

echo ">>> Updating modpack..."
mrpack-install \
  --server-dir "$MC_DIR" \
  --server-file server.jar \
  "$MRPACK_URL"

chown -R minecraft:minecraft "$MC_DIR"

echo ">>> Starting server..."
systemctl start minecraft

echo ">>> Done! Logs:"
journalctl -u minecraft -f
