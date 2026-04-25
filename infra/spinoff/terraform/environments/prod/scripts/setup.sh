#!/bin/bash
set -euo pipefail

LOG="/var/log/minecraft-setup.log"
exec > >(tee -a "$LOG") 2>&1

echo "=== Starting Minecraft server setup ==="

# --- Variables from Terraform ---
NEOFORGE_VERSION="${neoforge_version}"
MC_USER="minecraft"
MC_DIR="/opt/minecraft"
JAVA_VERSION="21"

# --- Update system ---
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

# --- Create user ---
echo ">>> Creating user $MC_USER..."
useradd -m -s /bin/bash "$MC_USER"

# --- Add minecraft user to sudoers ---
cat > /etc/sudoers.d/$MC_USER << 'EOF'
minecraft ALL=(ALL) NOPASSWD: ALL
EOF
chmod 440 /etc/sudoers.d/$MC_USER

# --- Copy authorized_keys to minecraft user ---
echo ">>> Copying authorized_keys to $MC_USER user..."
mkdir -p /home/"$MC_USER"/.ssh
cp /root/.ssh/authorized_keys /home/"$MC_USER"/.ssh/authorized_keys
chown -R "$MC_USER":"$MC_USER" /home/"$MC_USER"/.ssh
chmod 700 /home/"$MC_USER"/.ssh
chmod 600 /home/"$MC_USER"/.ssh/authorized_keys

# --- Create directory ---
mkdir -p "$MC_DIR"
chown "$MC_USER":"$MC_USER" "$MC_DIR"

# --- Install mrpack-install ---
echo ">>> Installing mrpack-install..."
MRPACK_INSTALL_VERSION="0.21.0-beta"
wget "https://github.com/nothub/mrpack-install/releases/download/v$${MRPACK_INSTALL_VERSION}/mrpack-install_$${MRPACK_INSTALL_VERSION}_linux_amd64.deb"
sudo apt install ./mrpack-install_$${MRPACK_INSTALL_VERSION}_linux_amd64.deb
rm mrpack-install_$${MRPACK_INSTALL_VERSION}_linux_amd64.deb

# --- Accept EULA ---
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

# --- Server icon ---
echo ">>> Downloading server-icon.png..."
wget -q "https://raw.githubusercontent.com/keelfy/lania/refs/heads/main/public/server-icon.png" \
  -O "$MC_DIR/server-icon.png"
chown "$MC_USER":"$MC_USER" "$MC_DIR/server-icon.png"

# --- Write start.sh ---
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

# --- Systemd service ---
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

echo "=== Setup complete ==="
touch /var/log/minecraft-setup-done
echo "Server IP will be available in Terraform output"
echo "Logs: journalctl -u minecraft -f"