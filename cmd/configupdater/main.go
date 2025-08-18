//go:build k8s
// +build k8s

package main

import (
	"context"
	"os"
	"sync"

	// configupdater "github.com/kloudmate/km-agent/configupdater"
	"github.com/kloudmate/km-agent/internal/config"
	"github.com/kloudmate/km-agent/internal/updater"
	cli "github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"go.uber.org/zap"
)

var (
	version = "0.1.0"
	commit  = "none"
)

func updaterFlags(cfg *config.K8sAgentConfig) []cli.Flag {
	return []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "api-key",
			Usage:       "API key for authentication",
			EnvVars:     []string{"KM_API_KEY"},
			Destination: &cfg.APIKey,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "collector-endpoint",
			Usage:       "OpenTelemetry exporter endpoint",
			Value:       "https://otel.kloudmate.com:4318",
			EnvVars:     []string{"KM_COLLECTOR_ENDPOINT"},
			Destination: &cfg.ExporterEndpoint,
		}),

		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "config-check-interval",
			Usage:       "Interval in seconds to check for config updates",
			Value:       "30s",
			EnvVars:     []string{"KM_CONFIG_CHECK_INTERVAL"},
			Destination: &cfg.ConfigCheckInterval,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "update-endpoint",
			Usage:       "API key for authentication",
			EnvVars:     []string{"KM_UPDATE_ENDPOINT"},
			Destination: &cfg.ConfigUpdateURL,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "kube-cluster-name",
			EnvVars:     []string{"KM_CLUSTER_NAME"},
			Destination: &cfg.ClusterName,
			Required:    true,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "kube-namespace",
			EnvVars:     []string{"KM_NAMESPACE"},
			Destination: &cfg.KubeNamespace,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "configmap-daemonset-name",
			EnvVars:     []string{"CONFIGMAP_DAEMONSET_NAME"},
			Destination: &cfg.ConfigmapDaemonsetName,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "configmap-deployment-name",
			EnvVars:     []string{"CONFIGMAP_DEPLOYMENT_NAME"},
			Destination: &cfg.ConfigmapDeploymentName,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "daemonset-name",
			EnvVars:     []string{"KM_DAEMONSET_NAME"},
			Destination: &cfg.DaemonSetName,
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "deployment-name",
			EnvVars:     []string{"KM_DEPLOYMENT_NAME"},
			Destination: &cfg.DeploymentName,
		}),
	}
}

func main() {
	var agentCfg config.K8sAgentConfig
	updaterflags := updaterFlags(&agentCfg)
	loggerConfig := zap.NewProductionConfig()
	logger, _ := loggerConfig.Build()

	app := &cli.App{
		Name:  "km-kube-updater",
		Usage: "Kloudmate's Kubernetes Agent Config Updater",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "check for updated config",
				Flags: updaterflags,
				Action: func(c *cli.Context) error {
					ctx, cancel := context.WithCancel(c.Context)
					defer cancel()

					logger.Sugar().Infow("kloudmate config updater info",
						"version", version,
						"commitSHA", commit,
					)
					logger.Info("loading InClusterConfig via service account.", zap.Any("config", agentCfg))
					kubeconfig, err := rest.InClusterConfig()
					if err != nil {
						return err
					}

					clientset, err := kubernetes.NewForConfig(kubeconfig)
					if err != nil {
						return err
					}

					kubeAgentConfig, err := config.NewKubeConfig(agentCfg, clientset, logger)
					if err != nil {
						logger.Fatal("failed to create kube agent config", zap.Error(err))
						return err
					}
					kubeUpdater := updater.NewKubeConfigUpdaterClient(kubeAgentConfig, logger.Sugar())
					kubeUpdater.SetConfigPath()

					var wg sync.WaitGroup
					errCh := make(chan error)
					wg.Add(1)
					go func() {
						defer wg.Done()
						kubeUpdater.StartConfigUpdateChecker(ctx)
					}()

					for err := range errCh {
						if err != nil {
							logger.Info("error while fetching config changes", zap.Error(err))
						}
					}
					close(kubeAgentConfig.StopCh)
					wg.Wait()
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		logger.Fatal("config updater cannot failed to start : ", zap.Error(err))
	}
}
