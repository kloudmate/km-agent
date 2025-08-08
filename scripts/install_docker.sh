#!/bin/bash

set -e

# Check Docker
if ! command -v docker &> /dev/null || ! docker --version &> /dev/null; then
  echo -e "\nDocker is not installed on the system"
  echo -e "\nPlease install docker first: https://docs.docker.com/engine/install/\n"
  exit 1
fi

DOCKER_SOCK_PATH="${DOCKER_SOCK_PATH:-/var/run/docker.sock}"
IMAGE_NAME="ghcr.io/kloudmate/km-agent:latest"


# Function to uninstall agent and remove container/image
uninstall_agent() {
    echo "🧹 Uninstalling km-agent..."

    if [ "$(docker ps -aq -f name=km-agent)" ]; then
        docker stop km-agent
        docker rm km-agent
        echo "✅ Container 'km-agent' stopped and removed."
    else
        echo "⚡ No running 'km-agent' container found."
    fi

    if docker image inspect $IMAGE_NAME > /dev/null 2>&1; then
        docker rmi $IMAGE_NAME
        echo "✅ Docker image removed."
    else
        echo "⚡ No '$IMAGE_NAME' image found."
    fi

    echo "✅ Uninstallation complete."
    exit 0
}

# --- Parse Arguments ---
if [[ "$1" == "uninstall" ]]; then
    uninstall_agent
fi

# Read from environment variables
KM_API_KEY="${KM_API_KEY}"
KM_COLLECTOR_ENDPOINT="${KM_COLLECTOR_ENDPOINT}"

# Validate inputs
if [ -z "$KM_API_KEY" ] || [ -z "$KM_COLLECTOR_ENDPOINT" ]; then
    echo "❌ Error: Both KM_API_KEY and KM_COLLECTOR_ENDPOINT must be provided as environment variables."
    echo "Usage:"
    echo "  KM_API_KEY=your_key KM_COLLECTOR_ENDPOINT=your_endpoint bash -c \"\$(curl -fsSL https://cdn.kloudmate.com/scripts/install_docker.sh)\""
    exit 1
fi

# Prompt for additional directories to monitor
ADDITIONAL_VOLUMES=""
read -p "📂 Do you want to monitor additional directories for logs? (y/n): " monitor_extra
if [[ "$monitor_extra" =~ ^[Yy]$ ]]; then
    while true; do
        read -r -p "📁 Enter the full path of the directory to monitor [e.g., /var/log/app] or leave empty to finish: " dir
        if [ -z "$dir" ]; then
            break
        elif [ -d "$dir" ]; then
            ADDITIONAL_VOLUMES+=" -v \"$dir:$dir:ro\""
        else
            echo "❌ Directory '$dir' does not exist. Try again."
        fi
    done
fi

echo "📥 Pulling Docker image: $IMAGE_NAME..."
docker pull $IMAGE_NAME

echo "🛑 Stopping and removing any existing 'km-agent' container..."
if [ "$(docker ps -aq -f name=km-agent)" ]; then
    docker stop km-agent
    docker rm km-agent
fi

echo "🚀 Running the 'km-agent' container..."
eval docker run -d \
  --privileged \
  --pid host \
  --restart always \
  --network host \
  --name km-agent-${KM_API_KEY:3:3} \
  -e KM_COLLECTOR_ENDPOINT="$KM_COLLECTOR_ENDPOINT" \
  -e KM_API_KEY="$KM_API_KEY" \
  -v "$DOCKER_SOCK_PATH":"$DOCKER_SOCK_PATH" \
  -v /var/log:/var/log \
  -v /var/lib/docker/containers:/var/lib/docker/containers:ro \
  $ADDITIONAL_VOLUMES \
  $IMAGE_NAME

echo "🎉 Setup complete! 'km-agent' is now running."
