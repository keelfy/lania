#!/bin/bash
set -euo pipefail

MC_USER="minecraft"
MC_DIR="/opt/minecraft"
MRPACK_FILE="server.mrpack"
MRPACK_URL="$1"
NEOFORGE_VERSION="$2"
RCON_PORT="$3"
RCON_PASSWORD="$4"

if [ -z "$MRPACK_URL" ]; then
  echo "Usage: ./update.sh <mrpack_url> <neoforge version> <rcon port> <rcon password>"
  exit 1
fi

echo ">>> Stopping server if running..."
sudo systemctl is-active --quiet minecraft && sudo systemctl stop minecraft || true

if [ -d "$MC_DIR/world" ]; then
  echo ">>> Clearing up mods folder..."
  rm -r "$MC_DIR/mods"
fi

echo ">>> Backing up world..."
if [ -d "$MC_DIR/world" ]; then
  cp -r "$MC_DIR/world" "$MC_DIR/world_backup_$(date +%Y%m%d_%H%M%S)"
  echo "World backed up"
else
  echo "No world found, skipping backup"
fi


# --- Accept EULA ---
echo "eula=true" > "$MC_DIR/eula.txt"
chown "$MC_USER":"$MC_USER" "$MC_DIR/eula.txt"

# --- server.properties ---
echo ">>> Writing server.properties..."
cat > "$MC_DIR/server.properties" << EOF
accepts-transfers=false
allow-flight=false
broadcast-console-to-ops=true
broadcast-rcon-to-ops=true
bug-report-link=
debug=false
difficulty=hard
enable-code-of-conduct=false
enable-jmx-monitoring=false
enable-query=false
enable-rcon=true
enable-status=true
enforce-secure-profile=true
enforce-whitelist=true
entity-broadcast-range-percentage=300
force-gamemode=false
function-permission-level=2
gamemode=survival
generate-structures=true
generator-settings={}
hardcore=false
hide-online-players=false
initial-disabled-packs=
initial-enabled-packs=vanilla
level-name=world
level-seed=zinc
level-type=minecraft\:normal
log-ips=true
management-server-allowed-origins=
management-server-enabled=false
management-server-host=localhost
management-server-port=0
management-server-secret=Pw4mZIYJgu0BBEm17ksNQl5G8WT7E00mInBTVfcB
management-server-tls-enabled=true
management-server-tls-keystore=
management-server-tls-keystore-password=
max-chained-neighbor-updates=1000000
max-players=20
max-tick-time=60000
max-world-size=29999984
motd=❤️ \u00A75\u00A7lʟᴀɴɪᴀ \u00A77- \u00A76sᴘɪɴᴏғғ sᴇrᴠᴇr
network-compression-threshold=256
online-mode=false
op-permission-level=4
pause-when-empty-seconds=-1
player-idle-timeout=0
prevent-proxy-connections=false
query.port=25565
rate-limit=0
rcon.password=$RCON_PASSWORD
rcon.port=$RCON_PORT
region-file-compression=deflate
require-resource-pack=false
resource-pack=
resource-pack-id=
resource-pack-prompt=
resource-pack-sha1=
server-ip=
server-port=25565
simulation-distance=8
spawn-protection=0
status-heartbeat-interval=0
sync-chunk-writes=true
text-filtering-config=
text-filtering-version=0
use-native-transport=true
view-distance=8
white-list=true
EOF
chown "$MC_USER":"$MC_USER" "$MC_DIR/server.properties"

# --- Server icon ---
echo ">>> Downloading server-icon.png..."
wget -q "https://raw.githubusercontent.com/keelfy/lania/refs/heads/main/public/server-icon.png" \
  -O "$MC_DIR/server-icon.png"
chown "$MC_USER":"$MC_USER" "$MC_DIR/server-icon.png"

# --- Write user_jvm_args.txt ---
echo ">>> Writing user_jvm_args.txt..."
echo "-Xms10G -Xmx16G " \
  "-XX:+UseG1GC " \
  "-XX:+ParallelRefProcEnabled " \
  "-XX:MaxGCPauseMillis=200 " \
  "-XX:+UnlockExperimentalVMOptions " \
  "-XX:+DisableExplicitGC " \
  "-XX:+AlwaysPreTouch " \
  "-XX:G1NewSizePercent=30 " \
  "-XX:G1MaxNewSizePercent=40 " \
  "-XX:G1HeapRegionSize=8M " \
  "-XX:G1ReservePercent=20 " \
  "-XX:G1HeapWastePercent=5 " \
  "-XX:G1MixedGCCountTarget=4 " \
  "-XX:InitiatingHeapOccupancyPercent=15 " \
  "-XX:G1MixedGCLiveThresholdPercent=90 " \
  "-XX:G1RSetUpdatingPauseTimePercent=5 " \
  "-XX:SurvivorRatio=32 " \
  "-XX:+PerfDisableSharedMem " \
  "-XX:MaxTenuringThreshold=1" > "$MC_DIR/user_jvm_args.txt"

echo ">>> Downloading server pack..."
wget "$MRPACK_URL" -O "$MC_DIR/$MRPACK_FILE"

echo ">>> Updating modpack..."
sudo mrpack-install --server-dir "$MC_DIR" --server-file server.jar "$MC_DIR/$MRPACK_FILE"

echo ">>> Configuring ownership..."
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
sudo java -jar "$INSTALLER" --install-server "$MC_DIR" 

sudo chmod +x "$MC_DIR/run.sh"
sudo chown "$MC_USER":"$MC_USER" "$MC_DIR/run.sh"

echo ">>> Cleaning up..."
rm -f "$INSTALLER"
rm -f "$INSTALLER.log"
rm -f "$INSTALLER"
rm -f "$MC_DIR/$MRPACK_FILE"

echo ">>> Starting server..."
sudo systemctl start "$MC_USER"

echo ">>> Server started. Done!"
echo "Logs: journalctl -u minecraft -f"
