#!/bin/bash

set -e

# Function to install Docker
install_docker() {
    echo "🚀 Checking if Docker is installed..."
    if ! command -v docker &> /dev/null; then
        echo "📦 Docker not found. Installing..."

        if [ -f /etc/debian_version ]; then
            echo "🔵 Debian-based system detected."
            sudo apt-get update
            sudo apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release
            curl -fsSL https://download.docker.com/linux/$(. /etc/os-release; echo "$ID")/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
            echo \
              "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/$(. /etc/os-release; echo "$ID") \
              $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
            sudo apt-get update
            sudo apt-get install -y docker-ce docker-ce-cli containerd.io
        elif [ -f /etc/redhat-release ]; then
            echo "🟠 RHEL-based system detected."
            sudo yum install -y yum-utils
            sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            sudo yum install -y docker-ce docker-ce-cli containerd.io
        else
            echo "❌ Unsupported OS. Please install Docker manually."
            exit 1
        fi

        echo "🚀 Starting and enabling Docker service..."
        sudo systemctl enable docker
        sudo systemctl start docker
        echo "✅ Docker installed successfully."
    else
        echo "✅ Docker is already installed."
    fi
}

# Function to uninstall agent and remove container/image
uninstall_agent() {
    echo "🧹 Uninstalling km-agent..."

    if [ "$(sudo docker ps -aq -f name=km-agent)" ]; then
        sudo docker stop km-agent
        sudo docker rm km-agent
        echo "✅ Container 'km-agent' stopped and removed."
    else
        echo "⚡ No running 'km-agent' container found."
    fi

    if sudo docker image inspect ghcr.io/kloudmate/km-agent:latest > /dev/null 2>&1; then
        sudo docker rmi ghcr.io/kloudmate/km-agent:latest
        echo "✅ Docker image removed."
    else
        echo "⚡ No 'km-agent' image found."
    fi

    echo "✅ Uninstallation complete."
    exit 0
}

# --- Parse Arguments ---
KM_API_KEY=""
KM_COLLECTOR_ENDPOINT=""

for arg in "$@"; do
  case $arg in
    KM_API_KEY=*)
      KM_API_KEY="${arg#*=}"
      ;;
    KM_COLLECTOR_ENDPOINT=*)
      KM_COLLECTOR_ENDPOINT="${arg#*=}"
      ;;
    uninstall)
      uninstall_agent
      ;;
    *)
      echo "❌ Unknown argument: $arg"
      echo "Usage:"
      echo "  bash script.sh KM_API_KEY=your_key KM_COLLECTOR_ENDPOINT=your_endpoint"
      echo "  or"
      echo "  bash script.sh uninstall"
      exit 1
      ;;
  esac
done

# Validate inputs
if [ -z "$KM_API_KEY" ] || [ -z "$KM_COLLECTOR_ENDPOINT" ]; then
    echo "❌ Error: Both KM_API_KEY and KM_COLLECTOR_ENDPOINT must be provided."
    exit 1
fi

# Docker image name
IMAGE_NAME="ghcr.io/kloudmate/km-agent:latest"

# --- Main Logic ---

install_docker

echo "📥 Pulling Docker image: $IMAGE_NAME..."
sudo docker pull $IMAGE_NAME

echo "🛑 Stopping and removing any existing 'km-agent' container..."
if [ "$(sudo docker ps -aq -f name=km-agent)" ]; then
    sudo docker stop km-agent
    sudo docker rm km-agent
fi

echo "🚀 Running the 'km-agent' container..."
sudo docker run -d \
  --privileged \
  --userns=host \
  --name km-agent \
  -e KM_COLLECTOR_ENDPOINT="$KM_COLLECTOR_ENDPOINT" \
  -e KM_API_KEY="$KM_API_KEY" \
  -v /:/hostfs:ro \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /etc/passwd:/etc/passwd:ro \
  $IMAGE_NAME

echo "🎉 Setup complete! 'km-agent' is now running."
