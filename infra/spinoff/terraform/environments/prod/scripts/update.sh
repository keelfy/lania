#!/bin/bash
set -euo pipefail

MC_USER="minecraft"
MC_DIR="/opt/minecraft"
MRPACK_URL="$1"
NEOFORGE_VERSION="$2"

if [ -z "$MRPACK_URL" ]; then
  echo "Usage: ./update.sh <mrpack_url> <neoforge version>"
  exit 1
fi

echo ">>> Stopping server if running..."
sudo systemctl is-active --quiet minecraft && sudo systemctl stop minecraft || true

echo ">>> Backing up world..."
if [ -d "$MC_DIR/world" ]; then
  cp -r "$MC_DIR/world" "$MC_DIR/world_backup_$(date +%Y%m%d_%H%M%S)"
  echo "World backed up"
else
  echo "No world found, skipping backup"
fi

echo ">>> Updating modpack..."
sudo mrpack-install \
  --server-dir "$MC_DIR" \
  --server-file server.jar \
  "$MRPACK_URL"

sudo chown -R "$MC_USER":"$MC_USER" "$MC_DIR"

echo ">>> Looking for NeoForge installer..."
INSTALLER="$(find "$MC_DIR" -maxdepth 1 -type f \
  -name 'neoforge-*-installer.jar' | sort -V | tail -n 1)"

if [ -z "$INSTALLER" ]; then
  echo "NeoForge installer not found in $MC_DIR"
  exit 1
fi

echo ">>> Found installer: $INSTALLER"
echo ">>> Installing NeoForge server files..."
sudo java -jar "$INSTALLER" \
  --install-server "$MC_DIR" \
  --launcher-jar "$MC_DIR/server.jar"

echo ">>> Cleaning up installer..."
rm -f "$INSTALLER"

echo ">>> Starting server..."
sudo systemctl start "$MC_USER"

echo ">>> Server started. Done!"
