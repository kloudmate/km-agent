# Deployment Configurations

This directory contains all the necessary configurations for deploying KmAgent to Kubernetes. It's structured into two main subdirectories:

* **`helm/`**: This directory holds KmAgent's [Helm chart](https://helm.sh/). Helm charts are used for defining, installing, and upgrading the complex Kubernetes based applications. [k8sagent](./helm/k8sagent/) subdirectory within `helm/` represents the agent Helm chart definitions.
* **`kubernetes/`**: This directory contains plain, standalone Kubernetes manifest files (e.g., `deployment.yaml`, `servicaaccount.yaml`, `cluster-role.yaml`,etc.). These are typically used for simpler deployments or resources not managed by Helm.

---