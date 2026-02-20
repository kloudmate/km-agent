#!/bin/bash
set -e

if [ -z "$KM_API_KEY" ] || [ -z "$KM_COLLECTOR_ENDPOINT" ]; then
  echo "❌ KM_API_KEY and KM_COLLECTOR_ENDPOINT must be set as environment variables"
  exit 1
fi

# --- Check/Install Required Tools ---
echo "🔧 Checking required tools..."

# install packages based on OS
install_pkg() {
  local pkg=$1
  if command -v apt-get &> /dev/null; then
    sudo apt-get update && sudo apt-get install -y "$pkg"
  elif command -v dnf &> /dev/null; then
    sudo dnf install -y "$pkg"
  elif command -v yum &> /dev/null; then
    sudo yum install -y "$pkg"
  else
    echo "❌ No package manager found to install $pkg"
    exit 1
  fi
}

for tool in curl wget systemctl; do
  if ! command -v $tool &> /dev/null; then
    echo "⚠️  Installing missing tool: $tool"
    install_pkg "$tool"
  fi
done

# jq is often not pre-installed on minimal AMIs
if ! command -v jq &> /dev/null; then
  echo "⚠️  Installing jq..."
  install_pkg "jq"
fi

# --- Fetch Latest Release Information ---
echo "🔎 Finding the latest KM Agent release from GitHub..."
API_URL="https://api.github.com/repos/kloudmate/km-agent/releases/latest"
LATEST_RELEASE_JSON=$(curl -s "$API_URL")

if [ -z "$LATEST_RELEASE_JSON" ] || [ "$(echo "$LATEST_RELEASE_JSON" | jq '.assets | length')" == "0" ]; then
    echo "❌ Error: Could not fetch latest release information from GitHub."
    exit 1
fi

VERSION=$(echo "$LATEST_RELEASE_JSON" | jq -r .tag_name)
echo "✅ Found latest version: $VERSION"

# --- Detect OS and Package Type ---
ARCH=$(uname -m)
PKG=""
PACKAGE_URL=""
INSTALL_CMD=""

if [ -f /etc/os-release ]; then
  . /etc/os-release
  case "$ID" in
    ubuntu|debian)
      PKG="deb"
      PACKAGE_URL=$(echo "$LATEST_RELEASE_JSON" | jq -r '.assets[] | select(.name | endswith(".deb")) | .browser_download_url')
      INSTALL_CMD="dpkg -i"
      ;;
    rhel|centos|rocky|almalinux|fedora|amzn|ol)  # Added amzn (Amazon Linux) and ol (Oracle Linux)
      PKG="rpm"
      PACKAGE_URL=$(echo "$LATEST_RELEASE_JSON" | jq -r '.assets[] | select(.name | endswith(".rpm")) | .browser_download_url')
      # Use yum/dnf to resolve dependencies instead of rpm -i
      if command -v dnf &> /dev/null; then
        INSTALL_CMD="dnf install -y"
      else
        INSTALL_CMD="yum install -y"
      fi
      ;;
    *)
      echo "❌ Error: Unsupported OS: $ID"
      exit 1
      ;;
  esac
else
  echo "❌ Error: Cannot detect operating system."
  exit 1
fi

if [ -z "$PACKAGE_URL" ] || [ "$PACKAGE_URL" == "null" ]; then
    echo "❌ Error: Could not find a .$PKG package in the latest release ($VERSION)."
    exit 1
fi

# --- Download and Install ---
TMP_PACKAGE="/tmp/kmagent.${PKG}"
echo "📥 Downloading KM Agent from $PACKAGE_URL ..."
wget -q "$PACKAGE_URL" -O "$TMP_PACKAGE"

echo "📦 Installing KM Agent..."
if [ "$PKG" = "deb" ]; then
  # For deb, we still use dpkg but handle dependencies with apt-get if needed
  if ! sudo KM_API_KEY="$KM_API_KEY" KM_COLLECTOR_ENDPOINT="$KM_COLLECTOR_ENDPOINT" dpkg -i "$TMP_PACKAGE"; then
    echo "⚠️  Fixing dependencies with apt-get..."
    sudo apt-get install -f -y
  fi
elif [ "$PKG" = "rpm" ]; then
  sudo KM_API_KEY="$KM_API_KEY" KM_COLLECTOR_ENDPOINT="$KM_COLLECTOR_ENDPOINT" $INSTALL_CMD "$TMP_PACKAGE"
fi

echo "🚀 Enabling and starting kmagent via systemd..."
sudo systemctl daemon-reexec
sudo systemctl enable kmagent
sudo systemctl restart kmagent

echo "✅ KM Agent installed and running as a systemd service."
echo "👉 To check status: sudo systemctl status kmagent"
echo " "
echo "🐛 Found a bug? Report to: support@kloudmate.com"
echo "   GitHub Issues: https://github.com/kloudmate/km-agent/issues"
exit 0