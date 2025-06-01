package kube

import (
	"fmt"
	"sync"

	"context"

	"github.com/kloudmate/km-agent/internal/shared"
	"github.com/kloudmate/km-agent/internal/updater"
	"go.opentelemetry.io/collector/otelcol"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

type KubeAgent struct {
	Cfg         *KubeAgentConfig
	Logger      *zap.SugaredLogger
	Collector   *otelcol.Collector
	K8sClient   *kubernetes.Clientset
	Updater     *updater.ConfigUpdater
	Errs        (chan error)
	collectorMu sync.Mutex
}

func NewKubeAgent() (*KubeAgent, error) {
	// ---------- Logging ----------
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}
	logger := zapLogger.Sugar()
	logger.Infow("bootstrapping kube agent")

	// ---------- Load YAML config ----------
	cfg, err := LoadKubeAgentConfig()
	if err != nil {
		logger.Errorw("failed to load agent config", "err", err)
		return nil, err
	}

	// ---------- Initialize Kubernetes client ----------
	k8sClient, err := initKubeClient(logger)
	if err != nil {
		logger.Errorw("failed to create k8s client", "err", err)
		return nil, err
	}

	/* ---------- config updater ----------
	TODO: implementation of KubeAgentConfig rather than config.Config for updater
	updaterCfg := updater.NewConfigUpdater(cfg, logger)
	*/

	// ---------- Create Kube agent ----------
	agent := &KubeAgent{
		Cfg:       cfg,
		Logger:    logger,
		K8sClient: k8sClient,
		Updater:   &updater.ConfigUpdater{},
		Errs:      make(chan error),
	}

	logger.Infow("kube agent initialized successfully")
	return agent, nil
}

func (km *KubeAgent) setupCollector(configPath string) error {
	collectorSettings := shared.CollectorInfoFactory(configPath)

	app, err := otelcol.NewCollector(collectorSettings)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/*
		TODO: Optional: could cancel later for hot reload
		go func() {
		 	time.Sleep(24 * time.Hour)
		 	cancel()
		}()
	*/

	return app.Run(ctx)
}

func (km *KubeAgent) StartOTelWithGeneratedConfig(config interface{}) error {
	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	tmpPath, err := writeTempOtelConfig(yamlBytes)
	if err != nil {
		return fmt.Errorf("temp config write error: %w", err)
	}

	return km.setupCollector(tmpPath)
}
