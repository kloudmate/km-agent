#!/bin/bash
set -e

if [ -z "$KM_API_KEY" ] || [ -z "$KM_COLLECTOR_ENDPOINT" ]; then
  echo "❌ KM_API_KEY and KM_COLLECTOR_ENDPOINT must be set as environment variables"
  exit 1
fi

REPO="kloudmate/km-agent"

for tool in curl wget systemctl jq; do
  if ! command -v $tool &> /dev/null; then
    echo "❌ Error: Required tool '$tool' is not installed. Please install it to continue."
    exit 1
  fi
done

# --- Fetch Latest Release Information ---
echo "🔎 Finding the latest KM Agent release from GitHub..."
API_URL="https://api.github.com/repos/$REPO/releases/latest"
LATEST_RELEASE_JSON=$(curl -s "$API_URL")

# Check if the API call was successful and returned assets.
if [ -z "$LATEST_RELEASE_JSON" ] || [ "$(echo "$LATEST_RELEASE_JSON" | jq '.assets | length')" == "0" ]; then
    echo "❌ Error: Could not fetch latest release information from GitHub."
    echo "   Please check the repository name and your network connection."
    exit 1
fi

VERSION=$(echo "$LATEST_RELEASE_JSON" | jq -r .tag_name)
echo "✅ Found latest version: $VERSION"


ARCH=$(uname -m)
PKG=""
PACKAGE_URL=""

if [ -f /etc/os-release ]; then
  . /etc/os-release
  case "$ID" in
    ubuntu|debian)
      PKG="deb"
      # Find the asset URL ending in .deb
      PACKAGE_URL=$(echo "$LATEST_RELEASE_JSON" | jq -r '.assets[] | select(.name | endswith(".deb")) | .browser_download_url')
      ;;
    rhel|centos|rocky|almalinux|fedora)
      PKG="rpm"
      # Find the asset URL ending in .rpm
      PACKAGE_URL=$(echo "$LATEST_RELEASE_JSON" | jq -r '.assets[] | select(.name | endswith(".rpm")) | .browser_download_url')
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

# Verify that we found a package URL for the detected OS.
if [ -z "$PACKAGE_URL" ] || [ "$PACKAGE_URL" == "null" ]; then
    echo "❌ Error: Could not find a .$PKG package in the latest release ($VERSION)."
    exit 1
fi


TMP_PACKAGE="/tmp/kmagent.${PKG}"
echo "📥 Downloading KM Agent from $PACKAGE_URL ..."
wget -q "$PACKAGE_URL" -O "$TMP_PACKAGE"


echo "📦 Installing KM Agent..."
if [ "$PKG" = "deb" ]; then
  sudo KM_API_KEY="$KM_API_KEY" KM_COLLECTOR_ENDPOINT="$KM_COLLECTOR_ENDPOINT" dpkg -i "$TMP_PACKAGE"
elif [ "$PKG" = "rpm" ]; then
  sudo KM_API_KEY="$KM_API_KEY" KM_COLLECTOR_ENDPOINT="$KM_COLLECTOR_ENDPOINT" rpm -i "$TMP_PACKAGE"
fi

echo "🚀 Enabling and starting kmagent via systemd..."
sudo systemctl daemon-reexec
sudo systemctl enable kmagent
sudo systemctl restart kmagent

echo "✅ KM Agent installed and running as a systemd service."
echo "👉 To check status: sudo systemctl status kmagent"
exit 0
