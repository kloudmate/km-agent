#!/bin/bash
set -e


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