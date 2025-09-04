# KM-Agent

<div align="center">

![KM-Agent Banner](docs/banner_km_agent.png)

[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Release](https://img.shields.io/github/release/kloudmate/km-agent.svg)](https://github.com/kloudmate/km-agent/releases)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/kloudmate/km-agent/actions)

[![GHCR Pulls](https://ghcr-badge.elias.eu.org/shield/kloudmate/km-agent/km-kube-agent)](https://ghcr-badge.elias.eu.org/shield/kloudmate/km-agent/km-kube-agent)

**KloudMate Agent for OpenTelemetry Auto Instrumentation**

*Simplifying OpenTelemetry adoption through automated deployment and remote configuration*

 • [Official Documentation](https://docs.kloudmate.com/kloudmate-agents) 

</div>


### Key Problems Solved

- **Complex Configuration**: Eliminates the steep learning curve of OpenTelemetry Collector configuration
- **Manual Installation**: Provides automated installation scripts for multiple environments
- **Configuration Management**: Enables remote configuration through a web interface without SSH access

## Features

- 🚀 **Automated Installation**: One-command deployment across Linux, Docker, and Kubernetes
- 🌐 **Remote Configuration**: Configure agents through a web interface without your target machine access
- 📊 **Lifecycle Management**: Comprehensive management of OpenTelemetry Collector
- 🔍 **Synthetic Monitoring**: Built-in health checks and monitoring capabilities
- 🎯 **Multi-Platform Support**: Native support for various deployment environments
- 📈 **Real-time Dashboards**: Unique agent identification for centralized monitoring

### Installation

Choose your environment and run the appropriate installation command:

#### Docker Installation
Docker agent is containerized version of the Agent that collect host level metrics (via `hostmetricreceiver`) and logs (via the volume mounts)
User can install the agent by running below script

```bash
KM_API_KEY="<YOUR_API_KEY>" KM_COLLECTOR_ENDPOINT="https://otel.kloudmate.com:4318" bash -c "$(curl -L https://cdn.kloudmate.com/scripts/install_docker.sh)"
```

#### Linux Installation
Similar to native OTel agent, agent supports both debian and Red Hat based systems.
User can install the agent via this automated bash script

```bash
KM_API_KEY="<YOUR_API_KEY>" KM_COLLECTOR_ENDPOINT="https://otel.kloudmate.com:4318" bash -c "$(curl -L https://cdn.kloudmate.com/scripts/install_linux.sh)"
```

Bash script should have various configurable arguments to configure the agent apart from API_KEY which is required for authentication at exporter. Each of the script should have corresponding uninstall command to remove the agent from the system.

#### Kubernetes Installation
The agent will run as DaemonSet as well as a Deployment in the cluster and add necessary components to monitor the nodes and pods
User can install the agent using below Helm based instructioins
```bash
helm repo add kloudmate https://kloudmate.github.io/km-agent
helm repo update
helm install kloudmate-release kloudmate/km-kube-agent --namespace km-agent --create-namespace \
--set API_KEY="<YOUR_API_KEY>" \n --set COLLECTOR_ENDPOINT="https://otel.kloudmate.com:4318" \
--set clusterName="<YOUR_CLUSTER_NAME>" \
--set monitoredNamespaces="<MONITORED_NS>"
```

#### Windows Installation
Download and run the Windows (.exe) installer from our [releases page](https://github.com/kloudmate/km-agent/releases).


## Supported Environments

![Supported Environments](docs/environments.png)

### Current Support
- ✅ **Linux** (Debian/Ubuntu, RHEL/CentOS)
- ✅ **Docker** (Host metrics and log collection)
- ✅ **Kubernetes** (via DaemonSet & Deployment)
- ✅ **Windows** (Windows Server 2016+)


### Architecture
Agent is installed as service on the host system/docker container/demonset on a k8s. It is done during installation process. The agent is responsible for managing the lifecycle of the Collector. The Agent is not an implementation of Collector, instead, it runs and manages lifecycle of existig OTel Collector.

![host_agent_lifecycle](/docs/lifecycle.png)

It is also primarily responsible for watching remote configuration (via REST endpoint) and pass on the configuration to Collector when changes has been detected. It has other functionalities such as synthetic monitoring that can be used to monitor the agent's status, various logs for monitoring purpose etc.

Each agent is uniquely identifyable so it can be used to build dashboard for the user to monitor the agents and configure them using a web interface.


### Kubernetes Agent Architecture

![K8s Agent Components](docs/km_agent_k8s.png)

The Kubernetes agent runs as a DaemonSet and includes:
- **Node Monitoring**: CPU, memory, disk, and network metrics
- **Pod Monitoring**: Container-level metrics and logs
- **Cluster Events**: Kubernetes events and resource monitoring
- **Service Discovery**: Automatic service endpoint detection

In future releases the agent can be installed in any of the following environments as well:
* Mac
* ECS
* Azure k8s


## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/kloudmate/km-agent.git
cd km-agent

# Build for Linux distribution
make build-linux-amd64
```

## Contributing

![Contributions Welcome](docs/contributions.png)

We welcome contributions that improve the quality, usability, and functionality of KM-Agent. Please read our contribution guidelines before getting started.

### How to Contribute

1. **Fork the Repository**
   ```bash
   git fork https://github.com/kloudmate/km-agent.git
   ```

2. **Create a Feature Branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

3. **Make Your Changes**
   - Follow our [coding standards](CONTRIBUTING.md#coding-standards)
   - Add tests for new functionality
   - Update documentation as needed

4. **Test Your Changes**
   ```bash
   make test
   make lint
   ```

5. **Submit a Pull Request**
   - Provide a clear description of your changes
   - Include any relevant issue numbers
   - Ensure all tests pass

### Reporting Issues

Before creating an issue, please:

- Check existing [issues](https://github.com/kloudmate/km-agent/issues)
- Use our issue templates
- Provide detailed reproduction steps
- Include environment information

### Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Community and Support

### Getting Help

- 📧 **Email**: support@kloudmate.com
- 🐛 **[Issues](https://github.com/kloudmate/km-agent/issues)** - Bug reports and feature requests
- 💻 **[Documentation](https://docs.kloudmate.com)**

### Community Resources

- 🌟 **[Dashboard Templates](https://github.com/kloudmate/dashboard-templates)**
- 📝 **[Blog Posts](https://blog.kloudmate.com)**
- 📱 **[Slack Community](https://kloudmate.slack.com)**


## License

This project is licensed under the [Apache License 2.0](LICENSE). See the LICENSE file for full details.

## Acknowledgments

-  **OpenTelemetry Community** - For the foundational observability framework
---

<div align="center">

**Made with 🧡 by the KloudMate Team**

[Website](https://kloudmate.com) • [Documentation](https://docs.kloudmate.com/kloudmate-agents) • [Community](https://github.com/kloudmate/km-agent/discussions) • [Support](mailto:support@kloudmate.com)

</div>