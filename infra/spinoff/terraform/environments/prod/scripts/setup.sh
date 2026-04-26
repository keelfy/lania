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
echo ">>> Ensuring user $MC_USER exists..."
if ! id "$MC_USER" >/dev/null 2>&1; then
  if [ -e "/home/$MC_USER" ] && [ ! -d "/home/$MC_USER" ]; then
    echo ">>> Removing invalid /home/$MC_USER path..."
    rm -f "/home/$MC_USER"
  fi
  useradd -m -s /bin/bash "$MC_USER"
fi

# --- Add minecraft user to sudoers ---
cat > /etc/sudoers.d/$MC_USER << 'EOF'
minecraft ALL=(ALL) NOPASSWD: ALL
EOF
chmod 440 /etc/sudoers.d/$MC_USER

# --- Copy authorized_keys to minecraft user ---
echo ">>> Copying authorized_keys to $MC_USER user..."
sudo mkdir -p /home/"$MC_USER"/.ssh
sudo cp /root/.ssh/authorized_keys /home/"$MC_USER"/.ssh/authorized_keys
sudo chown -R "$MC_USER":"$MC_USER" /home/"$MC_USER"/.ssh
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

# --- Install MCRCON ---
echo ">>> Installing mcrcon..."
MCRCON_VERSION="0.7.2"
curl -L -o /tmp/mcrcon.zip \
  "https://github.com/Tiiffi/mcrcon/releases/download/v$${MCRCON_VERSION}/mcrcon-$${MCRCON_VERSION}-linux-x86-64-static.zip"
unzip /tmp/mcrcon.zip -d /tmp/mcrcon
sudo install -m 755 "/tmp/mcrcon/mcrcon-$${MCRCON_VERSION}-linux-x86-64-static/mcrcon" /usr/local/bin/mcrcon
rm -rf /tmp/mcrcon /tmp/mcrcon.zip

# --- Systemd service ---
echo ">>> Creating systemd service..."
cat > /etc/systemd/system/minecraft.service << EOF
[Unit]
Description=Minecraft NeoForge Server
After=network.target

[Service]
User=$MC_USER
WorkingDirectory=$MC_DIR
ExecStart=$MC_DIR/run.sh
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
