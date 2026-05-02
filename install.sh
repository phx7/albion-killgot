#!/bin/bash
set -e

REPO="phx7/albion-killgot"
INSTALL_DIR="/opt/albion-killgot"
SERVICE_NAME="albion-killgot"

# Detect latest version
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
if [ -z "$VERSION" ]; then
  echo "Error: could not fetch latest version" >&2
  exit 1
fi

echo "Installing albion-killgot ${VERSION}..."

# Download binary
sudo mkdir -p "$INSTALL_DIR"
sudo curl -fsSL "https://github.com/${REPO}/releases/download/${VERSION}/bot-linux-amd64" \
  -o "${INSTALL_DIR}/bot"
sudo chmod +x "${INSTALL_DIR}/bot"

# Ask for token interactively if .env doesn't exist yet
if [ ! -f "${INSTALL_DIR}/.env" ]; then
  echo ""
  while true; do
    read -rp "Enter your Discord bot token: " DISCORD_TOKEN </dev/tty
    if [ -n "$DISCORD_TOKEN" ]; then
      break
    fi
    echo "Token cannot be empty."
  done
  printf "DISCORD_TOKEN=%s\n" "$DISCORD_TOKEN" | sudo tee "${INSTALL_DIR}/.env" > /dev/null
  echo "Saved to ${INSTALL_DIR}/.env"
fi

# Install systemd service
if command -v systemctl &>/dev/null; then
  sudo tee /etc/systemd/system/${SERVICE_NAME}.service > /dev/null <<EOF
[Unit]
Description=Albion Killgot Discord Bot
After=network.target

[Service]
ExecStart=${INSTALL_DIR}/bot
WorkingDirectory=${INSTALL_DIR}
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

  sudo systemctl daemon-reload
  sudo systemctl enable "${SERVICE_NAME}"
  echo ""
  echo "Installed as systemd service: ${SERVICE_NAME}"
  echo ""
  echo "Next steps:"
  echo "  sudo systemctl start ${SERVICE_NAME}"
  echo "  sudo journalctl -fu ${SERVICE_NAME}"
else
  echo ""
  echo "Installed to ${INSTALL_DIR}/bot"
  echo "Run: ${INSTALL_DIR}/bot"
fi
