//go:build k8s
// +build k8s

package k8sagent

import (
	"fmt"
	"os"
	"sync"
	"time"

	"context"

	"github.com/kloudmate/km-agent/internal/config"
	"github.com/kloudmate/km-agent/internal/updater"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

type K8sAgent struct {
	Cfg       *config.K8sAgentConfig
	Logger    *zap.SugaredLogger
	Collector *otelcol.Collector
	K8sClient *kubernetes.Clientset
	updater   *updater.K8sConfigUpdater

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

	// ---------- Load YAML config ----------
	cfg, err := config.LoadK8sAgentConfig()
	if err != nil {
		logger.Fatalw("failed to load agent config", "err", err)
		return nil, err
	}

	// ---------- Initialize Kubernetes client ----------
	k8sClient, err := initK8sClient(logger)
	if err != nil {
		logger.Errorw("failed to create k8s client", "err", err)
		return nil, err
	}

	updaterCfg := updater.NewK8sConfigUpdater(cfg, logger)

	// ---------- Create Kube agent ----------
	agent := &K8sAgent{
		Cfg:       cfg,
		Logger:    logger,
		K8sClient: k8sClient,
		updater:   updaterCfg,
		version:   version,
		stopCh:    make(chan struct{}),
	}

	logger.Infoln("kube agent initialized successfully")
	return agent, nil
}

// StartAgent first creates a otel config from agent config and then runs the agent
func (km *K8sAgent) StartAgent(ctx context.Context, cfg map[string]interface{}) error {

	return km.Start(ctx)
}

// Start runs the agent with otel config from the default path
func (a *K8sAgent) Start(ctx context.Context) error {
	a.Logger.Infoln("Starting collector agent...")

	// Start the initial collector instance
	if err := a.startInternalCollector(); err != nil {
		return fmt.Errorf("failed to start initial collector: %w", err)
	}

	// Start the configuration polling goroutine
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		a.runConfigUpdateChecker(ctx)
	}()

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

// runConfigUpdateChecker run ticker for performConfigCheck
func (a *K8sAgent) runConfigUpdateChecker(ctx context.Context) {
	if a.Cfg.ConfigUpdateURL == "" {
		a.Logger.Info("Config update URL not configured, skipping config update checks")
		return
	}
	ticker := time.NewTicker(time.Duration(a.Cfg.ConfigCheckInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.performConfigCheck(ctx); err != nil {
				a.Logger.Errorf("Periodic config check failed: %v", err)
			}
		case <-a.stopCh:
			a.Logger.Info("Config update checker stopping due to shutdown.")
			return
		case <-ctx.Done():
			a.Logger.Info("Config update checker stopping due to context cancellation.")
			return
		}
	}
}

// performConfigCheck checks remote server for new config and restart collector if required
func (a *K8sAgent) performConfigCheck(agentCtx context.Context) error {
	ctx, cancel := context.WithTimeout(agentCtx, 10*time.Second)
	defer cancel()

	a.Logger.Infoln("Checking for configuration updates...")

	a.collectorMu.Lock()
	params := updater.UpdateCheckerParams{
		Version: a.version,
	}
	if a.Collector != nil {
		params.CollectorStatus = "Running"
	} else {
		params.CollectorStatus = "Stopped"
	}
	a.collectorMu.Unlock()

	a.Logger.Debugf("Checking for updates with params: %+v", params)

	restart, newConfig, err := a.updater.CheckForUpdates(ctx, params)
	if err != nil {
		return fmt.Errorf("updater.CheckForUpdates failed: %w", err)
	}
	if newConfig != nil && restart {
		if err := a.UpdateConfigMap(newConfig); err != nil {
			return fmt.Errorf("failed to update config file: %w", err)
		}
		a.Logger.Infoln("Configuration change requires collector restart.")

		a.Stop()
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			if err := a.Start(ctx); err != nil {
				a.Logger.Errorf("failed to update config file: %w \n", err)
			} else {
				a.Logger.Infoln("Collector restarted successfully.")
			}
		}()
	} else {
		a.Logger.Infoln("No configuration change or restart required.")
	}
	return nil
}

func (a *K8sAgent) Stopch() {
	close(a.stopCh)
}

func (a *K8sAgent) UpdateConfigMap(cfg map[string]interface{}) error {

	yamlBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	// configDir := filepath.Dir(config.DefaultAgentConfigPath)
	// if err := os.MkdirAll(configDir, 0755); err != nil {
	// 	return fmt.Errorf("failed to create config directory: %w", err)
	// }

	configMaps := a.K8sClient.CoreV1().ConfigMaps(os.Getenv("POD_NAMESPACE"))

	_, err = configMaps.Update(context.TODO(), &corev1.ConfigMap{
		Data: map[string]string{"agent.yaml": string(yamlBytes)},
		ObjectMeta: v1.ObjectMeta{
			Name: os.Getenv("CONFIGMAP_NAME"),
		},
	}, v1.UpdateOptions{})

	if err != nil {
		a.Logger.Errorln(err)
	}

	// tempFile := config.DefaultAgentConfigPath + ".new"
	// if err := os.WriteFile(tempFile, yamlBytes, 0644); err != nil {
	// 	return fmt.Errorf("failed to write new config to temporary file: %w", err)
	// }

	// if err := os.Rename(tempFile, config.DefaultAgentConfigPath); err != nil {
	// 	return fmt.Errorf("failed to replace config file: %w", err)
	// }

	return nil
}
