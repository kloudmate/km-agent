//go:build k8s
// +build k8s

package k8sagent

import (
	"fmt"
	"sync"

	"context"

	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sAgent struct {
	// Cfg       *config.K8sAgentConfig
	Logger    *zap.SugaredLogger
	Collector *otelcol.Collector
	K8sClient *kubernetes.Clientset

	collectorMu     sync.Mutex
	wg              sync.WaitGroup
	collectorCtx    context.Context
	collectorCancel context.CancelFunc
	stopCh          chan struct{}
	version         string
}

func NewK8sAgent(version string) (*K8sAgent, error) {
	// ---------- Logging ----------
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	logger := zapLogger.Sugar()
	logger.Infow("bootstrapping kube agent")

	// ---------- Initialize Kubernetes client ----------

	kubecfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
	}
	logger.Infof("loaded cluster info from In-Cluster service account")
	k8sClient, err := kubernetes.NewForConfig(kubecfg)

	// ---------- Create Kube agent ----------
	agent := &K8sAgent{
		Logger:    logger,
		K8sClient: k8sClient,
		version:   version,
		stopCh:    make(chan struct{}),
	}

	logger.Infoln("kube agent initialized successfully")
	return agent, nil
}

// StartAgent first creates a otel config from agent config and then runs the agent
func (km *K8sAgent) StartAgent(ctx context.Context) error {

	return km.Start(ctx)
}

// Start runs the agent with otel config from the default path
func (a *K8sAgent) Start(ctx context.Context) error {
	a.Logger.Infoln("Starting collector agent...")

	// Start the initial collector instance
	if err := a.startInternalCollector(); err != nil {
		return fmt.Errorf("failed to start initial collector: %w", err)
	}
	a.Logger.Infoln("Collector agent started successfully.")
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
