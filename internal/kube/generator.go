package kube

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

func ParseKMAgentConfig(data []byte) (*KubeAgentConfig, error) {
	var cfg KubeAgentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse KMAgent config: %w", err)
	}
	return &cfg, nil
}

// GenerateCollectorConfig generates the otel config from the agentconfig provided to it.
func GenerateCollectorConfig(kcfg *KubeAgentConfig) (map[string]any, error) {
	if kcfg == nil {
		return nil, errors.New("nil config passed")
	}

	receivers := map[string]interface{}{}
	processors := map[string]interface{}{}
	exporters := map[string]interface{}{
		"logging": map[string]interface{}{
			"loglevel": "info",
		},
	}
	pipelines := map[string]interface{}{}

	collectionInterval := "30s"
	if kcfg.Monitoring.CollectionInterval != "" {
		collectionInterval = kcfg.Monitoring.CollectionInterval
	}

	// --- for Node Metrics ---
	if kcfg.Monitoring.Nodes.Enabled {
		receivers["hostmetrics"] = map[string]interface{}{
			"collection_interval": collectionInterval,
			"scrapers": map[string]interface{}{
				"cpu":     struct{}{},
				"memory":  struct{}{},
				"disk":    struct{}{},
				"network": struct{}{},
			},
		}

		pipelines["node_metrics"] = map[string]interface{}{
			"receivers":  []string{"hostmetrics"},
			"processors": []string{},
			"exporters":  []string{"logging"},
		}
	}

	// --- fore Pod Metrics ---
	if kcfg.Monitoring.Pods.Enabled {
		kubeletCfg := map[string]interface{}{
			"collection_interval":  collectionInterval,
			"auth_type":            "serviceAccount",
			"insecure_skip_verify": true,
			"metric_groups":        []string{"container", "pod"},
		}

		if kcfg.Monitoring.Pods.MonitorAllNamespaces {
			kubeletCfg["extra_metadata_labels"] = []string{"namespace"}
		} else {
			kubeletCfg["extra_metadata_labels"] = []string{"namespace"}
		}

		receivers["kubeletstats"] = kubeletCfg

		// for namespace filtering
		includeNs := kcfg.Monitoring.Pods.Namespaces.Include
		excludeNs := kcfg.Monitoring.Pods.Namespaces.Exclude

		filterRules := map[string]interface{}{}

		if len(includeNs) > 0 {
			filterRules["include"] = map[string]interface{}{
				"match_type": "strict",
				"resources": []map[string]interface{}{
					{
						"attributes": []map[string]interface{}{
							{
								"key":        "k8s.namespace.name",
								"values":     includeNs,
								"match_type": "strict",
							},
						},
					},
				},
			}
		}

		if len(excludeNs) > 0 {
			filterRules["exclude"] = map[string]interface{}{
				"match_type": "strict",
				"resources": []map[string]interface{}{
					{
						"attributes": []map[string]interface{}{
							{
								"key":        "k8s.namespace.name",
								"values":     excludeNs,
								"match_type": "strict",
							},
						},
					},
				},
			}
		}

		processors["filter/pod_ns"] = filterRules

		processors["k8sattributes"] = map[string]interface{}{
			"auth_type": "serviceAccount",
		}

		pipelines["pod_metrics"] = map[string]interface{}{
			"receivers":  []string{"kubeletstats"},
			"processors": []string{"k8sattributes", "filter/pod_ns"},
			"exporters":  []string{"logging"},
		}
	}

	// --- for Cluster Metrics ---
	if kcfg.Monitoring.Cluster.Enabled {
		receivers["k8s_cluster"] = map[string]interface{}{
			"collection_interval": collectionInterval,
		}

		pipelines["cluster_metrics"] = map[string]interface{}{
			"receivers":  []string{"k8s_cluster"},
			"processors": []string{},
			"exporters":  []string{"logging"},
		}
	}

	// --- for the Logs ---
	if kcfg.Monitoring.Logs.Enabled {
		logPaths := []string{}
		for _, src := range kcfg.Monitoring.Logs.Sources {
			switch src {
			case "kubelet_logs":
				logPaths = append(logPaths, "/var/log/kubelet.log")
			case "container_logs":
				logPaths = append(logPaths, "/var/log/pods/*/*.log")
			case "system_events":
				logPaths = append(logPaths, "/var/log/syslog")
			}
		}

		if len(logPaths) > 0 {
			receivers["filelog"] = map[string]interface{}{
				"include":  logPaths,
				"start_at": "beginning",
				"operators": []map[string]interface{}{
					{"type": "json_parser", "parse_from": "body"},
				},
			}

			processors["k8sattributes/logs"] = map[string]interface{}{
				"auth_type": "serviceAccount",
			}

			pipelines["logs"] = map[string]interface{}{
				"receivers":  []string{"filelog"},
				"processors": []string{"k8sattributes/logs"},
				"exporters":  []string{"logging"},
			}
		}
	}

	return map[string]any{
		"receivers":  receivers,
		"processors": processors,
		"exporters":  exporters,
		"service": map[string]any{
			"pipelines": pipelines,
		},
	}, nil
}
