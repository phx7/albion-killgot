#!/bin/bash
set -e

REPO="phx7/albion-killgot"
INSTALL_DIR="/opt/albion-killgot"
SERVICE_NAME="albion-killgot"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"

usage() {
  echo "Usage: $0 [install|update|uninstall]"
  echo ""
  echo "  install    Install the bot (default if already not installed)"
  echo "  update     Download latest release and restart the service"
  echo "  uninstall  Stop and remove the bot and its service"
  exit 1
}

fetch_latest_version() {
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)
  if [ -z "$VERSION" ]; then
    echo "Error: could not fetch latest version" >&2
    exit 1
  fi
}

download_binary() {
  echo "Downloading bot ${VERSION}..."
  sudo mkdir -p "$INSTALL_DIR"
  sudo curl -fsSL "https://github.com/${REPO}/releases/download/${VERSION}/bot-linux-amd64" \
    -o "${INSTALL_DIR}/bot"
  sudo chmod +x "${INSTALL_DIR}/bot"
  echo "Downloaded to ${INSTALL_DIR}/bot"
}

install_service() {
  sudo tee "$SERVICE_FILE" > /dev/null <<EOF
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
  sudo systemctl enable "$SERVICE_NAME"
}

cmd_install() {
  fetch_latest_version
  echo "Installing albion-killgot ${VERSION}..."

  download_binary

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

  if command -v systemctl &>/dev/null; then
    install_service
    sudo systemctl restart "$SERVICE_NAME"
    echo ""
    echo "Installed and started as systemd service: ${SERVICE_NAME}"
    echo "  sudo systemctl status ${SERVICE_NAME}"
    echo "  sudo journalctl -fu ${SERVICE_NAME}"
  else
    echo ""
    echo "Installed to ${INSTALL_DIR}/bot"
    echo "Run: ${INSTALL_DIR}/bot"
  fi
}

cmd_update() {
  fetch_latest_version
  echo "Updating albion-killgot to ${VERSION}..."

  if command -v systemctl &>/dev/null && systemctl is-active --quiet "$SERVICE_NAME"; then
    sudo systemctl stop "$SERVICE_NAME"
  fi

  download_binary

  if command -v systemctl &>/dev/null; then
    sudo systemctl restart "$SERVICE_NAME"
    echo "Service restarted."
    echo "  sudo journalctl -fu ${SERVICE_NAME}"
  else
    echo "Bot updated. Restart it manually."
  fi
}

cmd_uninstall() {
  echo "Uninstalling albion-killgot..."

  if command -v systemctl &>/dev/null; then
    sudo systemctl stop "$SERVICE_NAME" 2>/dev/null || true
    sudo systemctl disable "$SERVICE_NAME" 2>/dev/null || true
    sudo rm -f "$SERVICE_FILE"
    sudo systemctl daemon-reload
    echo "Systemd service removed."
  fi

  sudo rm -rf "$INSTALL_DIR"
  echo "Removed ${INSTALL_DIR}."
  echo "Uninstall complete."
}

# Auto-detect command if not provided
COMMAND="${1:-}"

if [ -z "$COMMAND" ]; then
  if [ -f "${INSTALL_DIR}/bot" ]; then
    COMMAND="update"
  else
    COMMAND="install"
  fi
fi

case "$COMMAND" in
  install)   cmd_install ;;
  update)    cmd_update ;;
  uninstall) cmd_uninstall ;;
  *)         usage ;;
esac
