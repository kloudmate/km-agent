#!/bin/bash
set -e

echo "🧹 KM Agent Uninstall Script"

# sometimes required
if [ "$EUID" -ne 0 ]; then 
  echo "⚠️  Not running as root. Some operations may fail."
fi

if ! command -v kmagent &>/dev/null && \
   ! dpkg -l kmagent &>/dev/null 2>&1 && \
   ! rpm -q kmagent &>/dev/null 2>&1; then
  echo "⚠️  KM Agent does not appear to be installed."
  read -p "Continue anyway to clean up residual files? [y/N] " -n 1 -r
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 0
  fi
fi

echo "🛑 Stopping KM Agent service..."
sudo systemctl stop kmagent 2>/dev/null || true
sudo systemctl disable kmagent 2>/dev/null || true

# Remove UNIT systemd service file
echo "🗑️  Removing systemd service..."
sudo rm -f /etc/systemd/system/kmagent.service
sudo rm -f /usr/lib/systemd/system/kmagent.service  # this is fallback bcoz some distros use this path
sudo systemctl daemon-reload
sudo systemctl reset-failed 2>/dev/null || true

# Clear journal logs for kmagent
echo "📝 Clearing journal logs..."
sudo journalctl --rotate 2>/dev/null || true
sudo journalctl --vacuum-time=1s --unit=kmagent 2>/dev/null || true

# Detect package manager and uninstall pkg properly
PKG_MANAGER=""
if command -v apt-get &>/dev/null && dpkg -l kmagent &>/dev/null 2>&1; then
  PKG_MANAGER="apt"
elif command -v dnf &>/dev/null && rpm -q kmagent &>/dev/null 2>&1; then
  PKG_MANAGER="dnf"
elif command -v yum &>/dev/null && rpm -q kmagent &>/dev/null 2>&1; then
  PKG_MANAGER="yum"
elif command -v zypper &>/dev/null && rpm -q kmagent &>/dev/null 2>&1; then
  PKG_MANAGER="zypper"
fi

if [ -n "$PKG_MANAGER" ]; then
  echo "📦 Removing package using $PKG_MANAGER..."
  case $PKG_MANAGER in
    apt)
      # purge removes config files too
      sudo apt-get purge -y kmagent 2>/dev/null || sudo dpkg -r kmagent
      sudo apt-get autoremove -y 2>/dev/null || true
      ;;
    dnf)
      sudo dnf remove -y kmagent
      ;;
    yum)
      sudo yum remove -y kmagent
      ;;
    zypper)
      sudo zypper remove -y kmagent
      ;;
  esac
else
  echo "⚠️  No package manager found or package not registered. Cleaning files manually..."
fi

# Remove binary and directories
echo "🗑️  Cleaning up files..."
sudo rm -f /usr/local/bin/kmagent
sudo rm -f /usr/bin/kmagent  # Alternative location
sudo rm -rf /etc/kmagent/
sudo rm -rf /var/log/kmagent/
sudo rm -rf /var/lib/kmagent/

# Clean up any temp files
sudo rm -f /tmp/kmagent.* /tmp/km-agent.* 2>/dev/null || true

# Clear shell history to remove potential API keys (recommended)
echo ""
echo "🔒 Note: Consider running 'history -c && exit' to clear shell history"
echo "   if API keys were visible in environment variables."
echo ""
echo "✅ KM Agent has been uninstalled."
echo "   Remaining traces (if any):"
echo "   - Systemd: systemctl list-units | grep kmagent"
echo "   - Files: find / -name '*kmagent*' 2>/dev/null"
# Bug reporting information
echo ""
echo "🐛 Found a bug? Report to: support@kloudmate.com"
echo "   GitHub Issues: https://github.com/kloudmate/km-agent/issues"
exit 0