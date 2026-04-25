#!/bin/bash
set -euo pipefail

LOG="/var/log/minecraft-setup.log"
exec > >(tee -a "$LOG") 2>&1

echo "=== Starting Minecraft server setup ==="

# --- Переменные из Terraform ---
NEOFORGE_VERSION="${neoforge_version}"
MRPACK_URL="${mrpack_url}"
MC_USER="minecraft"
MC_DIR="/opt/minecraft"
JAVA_VERSION="21"

# --- Обновление системы ---
echo ">>> Updating system..."
apt-get update -y
apt-get upgrade -y
apt-get install -y \
  curl \
  wget \
  unzip \
  git \
  screen \
  ufw \
  openjdk-$${JAVA_VERSION}-jre-headless

# --- Создание пользователя ---
echo ">>> Creating user $MC_USER..."
useradd -m -s /bin/bash "$MC_USER"

# --- Создание директории ---
mkdir -p "$MC_DIR"
chown "$MC_USER":"$MC_USER" "$MC_DIR"

# --- Установка mrpack-install ---
echo ">>> Installing mrpack-install..."
MRPACK_INSTALL_VERSION="0.6.0"
wget -q "https://github.com/nothub/mrpack-install/releases/download/v$${MRPACK_INSTALL_VERSION}/mrpack-install-linux-amd64" \
  -O /usr/local/bin/mrpack-install
chmod +x /usr/local/bin/mrpack-install

# --- Установка NeoForge через mrpack-install ---
echo ">>> Installing modpack from $MRPACK_URL..."
sudo -u "$MC_USER" mrpack-install \
  --server-dir "$MC_DIR" \
  --server-file server.jar \
  "$MRPACK_URL"

# --- Принятие EULA ---
echo "eula=true" > "$MC_DIR/eula.txt"
chown "$MC_USER":"$MC_USER" "$MC_DIR/eula.txt"

# --- server.properties ---
echo ">>> Writing server.properties..."
cat > "$MC_DIR/server.properties" << 'EOF'
server-port=25565
motd=❤️ \u00A75\u00A7lʟᴀɴɪᴀ \u00A77- \u00A76sᴘɪɴᴏғғ sᴇrᴠᴇr
white-list=true
enforce-whitelist=true
online-mode=false
max-players=20
view-distance=8
simulation-distance=8
difficulty=hard
gamemode=survival
EOF
chown "$MC_USER":"$MC_USER" "$MC_DIR/server.properties"

# --- Скрипт запуска ---
echo ">>> Writing start.sh..."
cat > "$MC_DIR/start.sh" << 'EOF'
#!/bin/bash
cd /opt/minecraft
exec java \
  -Xms10G -Xmx16G \
  -XX:+UseG1GC \
  -XX:+ParallelRefProcEnabled \
  -XX:MaxGCPauseMillis=200 \
  -XX:+UnlockExperimentalVMOptions \
  -XX:+DisableExplicitGC \
  -XX:+AlwaysPreTouch \
  -XX:G1NewSizePercent=30 \
  -XX:G1MaxNewSizePercent=40 \
  -XX:G1HeapRegionSize=8M \
  -XX:G1ReservePercent=20 \
  -XX:G1HeapWastePercent=5 \
  -XX:G1MixedGCCountTarget=4 \
  -XX:InitiatingHeapOccupancyPercent=15 \
  -XX:G1MixedGCLiveThresholdPercent=90 \
  -XX:G1RSetUpdatingPauseTimePercent=5 \
  -XX:SurvivorRatio=32 \
  -XX:+PerfDisableSharedMem \
  -XX:MaxTenuringThreshold=1 \
  -jar server.jar nogui
EOF
chmod +x "$MC_DIR/start.sh"
chown "$MC_USER":"$MC_USER" "$MC_DIR/start.sh"

# --- Systemd сервис ---
echo ">>> Creating systemd service..."
cat > /etc/systemd/system/minecraft.service << EOF
[Unit]
Description=Minecraft NeoForge Server
After=network.target

[Service]
User=$MC_USER
WorkingDirectory=$MC_DIR
ExecStart=$MC_DIR/start.sh
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable minecraft
systemctl start minecraft

echo "=== Setup complete ==="
echo "Server IP will be available in Terraform output"
echo "Logs: journalctl -u minecraft -f"
