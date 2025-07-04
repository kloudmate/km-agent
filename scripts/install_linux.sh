#!/bin/bash
set -e

# 🧾 Usage
usage() {
  echo "Usage:"
  echo "  --install     Install and start the KM Agent"
  echo "  --uninstall   Uninstall the KM Agent"
  exit 1
}

# 🧹 Uninstall logic
uninstall_kmagent() {
  echo "🧹 Stopping and removing KM Agent..."

  sudo systemctl stop kmagent || true
  sudo systemctl disable kmagent || true
  sudo rm -f /etc/systemd/system/kmagent.service
  sudo systemctl daemon-reload
  sudo rm -f /usr/local/bin/kmagent
  sudo rm -rf /etc/kmagent/ /var/log/kmagent/ /var/lib/kmagent/

  if command -v dpkg &>/dev/null && dpkg -l | grep -q kmagent; then
    sudo dpkg -r kmagent
  elif command -v rpm &>/dev/null && rpm -q kmagent &>/dev/null; then
    sudo rpm -e kmagent
  fi

  echo "✅ KM Agent uninstalled."
  exit 0
}

# 📦 Install logic
install_kmagent() {
  if [ -z "$KM_API_KEY" ] || [ -z "$KM_COLLECTOR_ENDPOINT" ]; then
    echo "❌ KM_API_KEY and KM_COLLECTOR_ENDPOINT must be set as environment variables"
    exit 1
  fi

  ARCH=$(uname -m)
  PKG=""
  PACKAGE_URL=""

  if [ -f /etc/os-release ]; then
    . /etc/os-release
    case "$ID" in
      ubuntu|debian)
        PKG="deb"
        PACKAGE_URL="https://github.com/kloudmate/km-agent/releases/download/1.0.0/kmagent_1.0.0_amd64.deb"
        ;;
      rhel|centos|rocky|almalinux|fedora)
        PKG="rpm"
        PACKAGE_URL="https://github.com/kloudmate/km-agent/releases/download/1.0.0/kmagent-1.0.0-1.x86_64.rpm"
        ;;
      *)
        echo "❌ Unsupported OS: $ID"
        exit 1
        ;;
    esac
  else
    echo "❌ Cannot detect OS"
    exit 1
  fi

  for tool in curl wget systemctl; do
    if ! command -v $tool &>/dev/null; then
      echo "❌ Required tool '$tool' is not installed."
      exit 1
    fi
  done

  TMP_PACKAGE="/tmp/kmagent.${PKG}"
  echo "📥 Downloading KM Agent from $PACKAGE_URL ..."
  wget -q "$PACKAGE_URL" -O "$TMP_PACKAGE"

  KM_UPDATE_ENDPOINT="${KM_UPDATE_ENDPOINT:-https://api.kloudmate.dev/agents/config-check}"

  echo "📦 Installing KM Agent..."
  if [ "$PKG" = "deb" ]; then
    sudo KM_API_KEY="$KM_API_KEY" KM_COLLECTOR_ENDPOINT="$KM_COLLECTOR_ENDPOINT" KM_UPDATE_ENDPOINT="$KM_UPDATE_ENDPOINT" dpkg -i "$TMP_PACKAGE"
  elif [ "$PKG" = "rpm" ]; then
    sudo KM_API_KEY="$KM_API_KEY" KM_COLLECTOR_ENDPOINT="$KM_COLLECTOR_ENDPOINT" KM_UPDATE_ENDPOINT="$KM_UPDATE_ENDPOINT" rpm -i "$TMP_PACKAGE"
  fi

  echo "🚀 Enabling and starting kmagent via systemd..."
  sudo systemctl daemon-reexec
  sudo systemctl enable kmagent
  sudo systemctl restart kmagent

  echo "✅ KM Agent installed and running as a systemd service."
  echo "👉 To check status: sudo systemctl status kmagent"
  exit 0
}

# 🚦 Main logic — using $0 directly
case "$0" in
  --install)
    install_kmagent
    ;;
  --uninstall)
    uninstall_kmagent
    ;;
  *)
    usage
    ;;
esac
