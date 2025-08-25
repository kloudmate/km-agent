//go:build k8s
// +build k8s

package k8sagent

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"context"

	"github.com/kloudmate/km-agent/internal/shared"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// K8sConfig holds all configuration values from environment variables
type K8sConfig struct {
	APIKey              string `env:"KM_API_KEY"`
	CollectorEndpoint   string `env:"KM_COLLECTOR_ENDPOINT"`
	ConfigCheckInterval string `env:"KM_CONFIG_CHECK_INTERVAL"`
	DeploymentMode      string `env:"DEPLOYMENT_MODE"`
	ConfigMapName       string `env:"CONFIGMAP_NAME"`
	PodNamespace        string `env:"POD_NAMESPACE"`
}

type K8sAgent struct {
	Cfg       *K8sConfig
	Logger    *zap.SugaredLogger
	Collector *otelcol.Collector
	K8sClient *kubernetes.Clientset

	collectorMu     sync.Mutex
	wg              sync.WaitGroup
	collectorCtx    context.Context
	collectorCancel context.CancelFunc
	stopCh          chan struct{}
	AgentInfo       AgentInfo
}

type AgentInfo struct {
	Version          string
	CommitSHA        string
	CollectorVersion string
}

func NewK8sAgent(info *AgentInfo) (*K8sAgent, error) {
	// ---------- Logging ----------
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	logger := zapLogger.Sugar()
	logger.Infow("bootstrapping kube agent")

	cfg := NewK8sConfig()

	// ---------- Initialize Kubernetes client ----------

	kubecfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}
	logger.Infof("loaded cluster info from In-Cluster service account")
	k8sClient, err := kubernetes.NewForConfig(kubecfg)

	// ---------- Create Kube agent ----------
	agent := &K8sAgent{
		Cfg:       cfg,
		Logger:    logger,
		K8sClient: k8sClient,
		AgentInfo: *info,
		stopCh:    make(chan struct{}),
	}
	agent.AgentInfo.setEnvForAgentVersion()
	agent.AgentInfo.CollectorVersion = shared.GetCollectorVersion()
	logger.Infoln("kube agent initialized successfully")
	return agent, nil
}

// StartAgent first creates a otel config from agent config and then runs the agent
func (km *K8sAgent) StartAgent(ctx context.Context) error {
	km.Logger.Infow("kloudmate kubernetes agent info",
		"version", km.AgentInfo.Version,
		"commitSHA", km.AgentInfo.CommitSHA,
		"collector-version", km.AgentInfo.CollectorVersion,
	)
	return km.Start(ctx)
}

// Start runs the agent with otel config from the default path
func (a *K8sAgent) Start(ctx context.Context) error {
	a.Logger.Infoln("Starting collector agent...")

	// Start the initial collector instance
	if err := a.startInternalCollector(); err != nil {
		return fmt.Errorf("failed to start initial collector: %w", err)
	} else {
		a.Logger.Infoln("collector agent started successfully.")
	}
	return nil
}

// Stop stops the underlying collector
func (a *K8sAgent) Stop() {
	a.Logger.Infoln("Stopping collector agent...")

	// Signal the polling goroutine to stop
	close(a.stopCh)
	a.wg.Wait()

	// Stop the collector instance
	a.stopInternalCollector()

	a.Logger.Infoln("Collector agent stopped.")
}

func (a *K8sAgent) Stopch() {
	close(a.stopCh)
}

func (a *K8sAgent) AwaitShutdown() {
	a.wg.Wait()
}

func NewK8sConfig() *K8sConfig {
	config := &K8sConfig{
		ConfigCheckInterval: os.Getenv(""),
		APIKey:              os.Getenv("KM_API_KEY"),
		CollectorEndpoint:   os.Getenv("KM_COLLECTOR_ENDPOINT"),
		ConfigMapName:       os.Getenv("CONFIGMAP_NAME"),
		DeploymentMode:      os.Getenv("DEPLOYMENT_MODE"),
		PodNamespace:        os.Getenv("POD_NAMESPACE"),
	}

	if strings.ToUpper(config.DeploymentMode) == "DAEMONSET" {
		config.DeploymentMode = "DAEMONSET"
	} else {
		config.DeploymentMode = "DEPLOYMENT"

	}
	return config
}

func (c *K8sConfig) Validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("KM_API_KEY is required")
	}
	if c.APIKey == "" {
		return fmt.Errorf("KM_COLLECTOR_ENDPOINT is required")
	}
	if c.ConfigMapName == "" {
		return fmt.Errorf("CONFIGMAP_NAME is required")
	}
	if c.PodNamespace == "" {
		return fmt.Errorf("POD_NAMESPACE is required")
	}
	return nil
}

func (c *K8sAgent) otelConfigPath() string {
	daemonsetURI := "/etc/kmagent/agent-daemonset.yaml"
	deploymentURI := "/etc/kmagent/agent-deployment.yaml"
	if c.Cfg.DeploymentMode == "DAEMONSET" {
		return daemonsetURI
	} else {
		return deploymentURI
	}
}

// setEnvForAgentVersion sets agent version on env this gets later used by otel processor to inject agent version
func (r *AgentInfo) setEnvForAgentVersion() {
	os.Setenv("KM_AGENT_VERSION", r.Version)
}
