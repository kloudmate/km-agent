# km-agent
KloudMate Agent for OpenTelemetry auto instrumentation

Purpose of KloudMate Agent is to auto instrument host system with OTel collector. This is a wrapper on top of Otel Collector for ease of deployment and management.

Right now, user has to go through various documents to understand and configure otel collector depending on components to be monitored. This is fairly complex process for someone new to OpenTelemetry. There is step learning curve which becomes a barrier to adopting OpenTelemetry. This custom KM Agent will solve two problems: 

1. Ease of installation using automated installation script (Bash script/Windows Installer).
2. Remote configuration of the collector. User can configure the agent from a web interface without having to login into the host system.

### Installation Script
Depending on type of environment, there should be different scripts (Linux, Docker, k8s etc). User will copy the command and execute on the target system to initiate installation.

For example,
```
API_KEY="<API_KEY>" bash -c "$(curl -L https://cdn.kloudmate.com/scripts/docker-agent.sh)"
```
Bash script should have various configurable arguments to configure the agent apart from API_KEY which is required for authentication at exporter. Each of the script should have corresponding uninstall command to remove the agent from the system.

### Agent
Agent is installed as service on the host system/docker container/demonset on a k8s. it is done during installation process. The agent is responsable for managing the lifecycle of the Collector. The Agent is not implimentaton of Collector, instead, it runs and manages lifecycle of existig Otel Collector.

It is primarily responsible for watching remote configuration (via REST endpoint) and pass on the configuration to Collector when changes has been detected. It has other functionalities such as sending heartbeat metric to external API which can be use to monitor the agent status, various logs for monitoring purpose etc.

Each agent should be uniquely identifyable so we can build dashboard for the user to monitor the agents and configure them using a web interface.

The agent can be installed in any of the following environments
* Docker
* Kubernetes 
* Linux
* Windows
* Mac
* ECS
* Azure k8s
* etc..

**Docker**

Docker agent is containerized version of the Agent that collect host level metrics (hostmetricreceiver) and logs (via volume mount)
User should be able to install this via automated bash script

**Kubernetes**

The agent will run as DaemonSet in the cluster and add necessary components to monitor the nodes and pods
User should be able to install this via automated bash script (bash/.bat/Helm)

**Linux**

Similar to native Otel agent, should support both debian and Red Hat based systems
User should be able to install this via automated bash script

**Windows**

Installation via Windows installer
