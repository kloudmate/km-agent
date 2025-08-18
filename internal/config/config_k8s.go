//go:build k8s
// +build k8s

package config

import (
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

const (
	// DefaultConfigmapMountPath = "/etc/kmagent/agent.yaml"

	EnvAPIKey              = "KM_API_KEY"
	EnvAgentConfig         = "KM_AGENT_CONFIG"
	EnvExporterEndpoint    = "KM_COLLECTOR_ENDPOINT"
	EnvConfigCheckInterval = "KM_CONFIG_CHECK_INTERVAL"
	EnvUpdateEndpoint      = "KM_UPDATE_ENDPOINT"
)

type K8sAgentConfig struct {
	Logger    *zap.SugaredLogger
	K8sClient *kubernetes.Clientset
	StopCh    chan struct{}
	Version   string

	OtelCollectorConfig map[string]interface{}
	ExporterEndpoint    string
	ConfigUpdateURL     string
	APIKey              string
	ConfigCheckInterval string
	// Kubernetes specific
	KubeNamespace           string
	ClusterName             string
	DeploymentMode          string
	ConfigmapDaemonsetName  string
	ConfigmapDeploymentName string
	DaemonSetName           string
	DeploymentName          string
}

func NewKubeConfig(cfg K8sAgentConfig, clientset *kubernetes.Clientset, logger *zap.Logger) (*K8sAgentConfig, error) {

	agent := &K8sAgentConfig{
		Logger:                  logger.Sugar(),
		K8sClient:               clientset,
		ExporterEndpoint:        cfg.ExporterEndpoint,
		ConfigUpdateURL:         GetAgentConfigUpdaterURL(cfg.ExporterEndpoint),
		APIKey:                  cfg.APIKey,
		ConfigCheckInterval:     cfg.ConfigCheckInterval,
		KubeNamespace:           cfg.KubeNamespace,
		ClusterName:             cfg.ClusterName,
		DaemonSetName:           cfg.DaemonSetName,
		ConfigmapDaemonsetName:  cfg.ConfigmapDaemonsetName,
		ConfigmapDeploymentName: cfg.ConfigmapDeploymentName,
	}

	agent.Logger.Infoln("kube updater initialized successfully")
	return agent, nil
}
