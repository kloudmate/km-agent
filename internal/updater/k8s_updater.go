package updater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/kloudmate/km-agent/internal/config"
	"github.com/kloudmate/km-agent/internal/instrumentation"
	"github.com/kloudmate/km-agent/internal/shared"
	"github.com/kloudmate/km-agent/rpc"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// ConfigUpdater handles configuration updates from a remote API
type K8sConfigUpdater struct {
	cfg        *config.K8sAgentConfig
	logger     *zap.SugaredLogger
	client     *http.Client
	configPath string
}

type K8sUpdateCheckerParams struct {
	Version          string
	CollectorVersion string
	CollectorStatus  string
	APMData          []APMConfig
}

type APMConfig struct {
	Namespace  string `json:"namespace"`
	Deployment string `json:"deployment"`
	Kind       string `json:"kind"`
	Enabled    bool   `json:"enabled"`
	Language   string `json:"language"`
}

type K8sOtelConfigs struct {
	DaemonSetConfig  map[string]interface{} `json:"daemonset_config"`
	DeploymentConfig map[string]interface{} `json:"deployment_config"`
}

type K8sApmConfig struct {
	APMEnabled  bool        `json:"apm_enabled"`
	APMSettings []APMConfig `json:"apm_settings"`
}

// ConfigUpdateResponse represents the response from the config update API
type K8sConfigUpdateResponse struct {
	RestartRequired bool           `json:"restart_required"`
	K8sAPIConfigs   K8sOtelConfigs `json:"config"`
	K8s             K8sApmConfig   `json:"k8s"`
}

// NewK8sConfigUpdater creates a new config updater
func NewKubeConfigUpdaterClient(cfg *config.K8sAgentConfig, logger *zap.SugaredLogger) *K8sConfigUpdater {

	return &K8sConfigUpdater{
		cfg:    cfg,
		logger: logger,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				DialContext:           (&net.Dialer{Timeout: 5 * time.Second}).DialContext,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 5 * time.Second,
			},
		},
	}
}

// CheckForUpdates checks for configuration updates from the remote API
func (u *K8sConfigUpdater) CheckForUpdatesK8s(ctx context.Context, p K8sUpdateCheckerParams) (updateResp K8sConfigUpdateResponse, err error) {

	// Create the request
	data := map[string]interface{}{
		"architecture":      runtime.GOARCH,
		"hostname":          u.cfg.ClusterName,
		"platform":          "k8s",
		"k8s_deployments":   p.APMData,
		"collector_version": shared.GetCollectorVersion(),
		"agent_version":     p.Version,
		"collector_status":  p.CollectorStatus,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel() // Ensure context resources are freed

	req, err := http.NewRequestWithContext(reqCtx, "POST", u.cfg.ConfigUpdateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return K8sConfigUpdateResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Add API key if configured
	if u.cfg.APIKey != "" {
		req.Header.Set("Authorization", u.cfg.APIKey)
	}

	resp, respErr := u.client.Do(req)

	if respErr != nil {
		return K8sConfigUpdateResponse{}, fmt.Errorf("failed to fetch config updates after retries: %w", respErr)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return K8sConfigUpdateResponse{}, fmt.Errorf("config update API returned non-OK status: %d, body: %s", resp.StatusCode, body)
	}

	// Parse response
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		return K8sConfigUpdateResponse{}, fmt.Errorf("failed to decode config update response: %w", err)
	}

	return updateResp, nil
}

// StartConfigUpdateChecker run ticker for performConfigCheck
func (a *K8sConfigUpdater) StartConfigUpdateChecker(ctx context.Context) {
	if a.cfg.ConfigUpdateURL == "" {
		a.logger.Info("Config update URL not configured, skipping config update checks")
		return
	}
	parsedTime, err := time.ParseDuration(a.cfg.ConfigCheckInterval)
	if err != nil {
		a.logger.Info("Config update URL parse error, falling back to default value")
		parsedTime = time.Duration(time.Second * 30)
	}
	ticker := time.NewTicker(parsedTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := a.performConfigCheck(ctx); err != nil {
				a.logger.Errorf("Periodic config check failed: %v", err)
			}
		case <-a.cfg.StopCh:
			a.logger.Info("Config update checker stopping due to shutdown.")
			return
		case <-ctx.Done():
			a.logger.Info("Config update checker stopping due to context cancellation.")
			return
		}
	}
}

// performConfigCheck checks remote server for new config and restart collector if required
func (a *K8sConfigUpdater) performConfigCheck(agentCtx context.Context) error {
	ctx, cancel := context.WithTimeout(agentCtx, 15*time.Second)
	defer cancel()

	a.logger.Infoln("Checking for configuration updates...")
	apmData := []APMConfig{}
	results := rpc.GetDetectionResults()
	a.logger.Infoln("available apps for instrumentation : %d", len(results))
	for _, info := range results {
		apmData = append(apmData, APMConfig{
			Namespace:  info.Namespace,
			Deployment: info.DeploymentName,
			Kind:       info.Kind,
			Language:   info.Language,
			Enabled:    info.Enabled,
		})
	}
	params := K8sUpdateCheckerParams{
		Version:          a.cfg.Version,
		CollectorVersion: shared.GetCollectorVersion(),
		CollectorStatus:  "Running",
		APMData:          apmData,
	}

	a.logger.Debugf("Checking for updates with params: %+v", params)

	updateResp, err := a.CheckForUpdatesK8s(ctx, params)
	if err != nil {
		return fmt.Errorf("updater.CheckForUpdates failed: %w", err)
	}
	if updateResp.K8sAPIConfigs.DaemonSetConfig != nil && updateResp.K8sAPIConfigs.DeploymentConfig != nil && updateResp.RestartRequired {

		if err := a.UpdateConfigMap(updateResp.K8sAPIConfigs.DaemonSetConfig, updateResp.K8sAPIConfigs.DeploymentConfig); err != nil {
			return fmt.Errorf("failed to update configMap: %w", err)
		}
		a.logger.Infoln("triggering rollout restart.")

		if err = a.triggerDaemonSetRollout(agentCtx); err != nil {
			a.logger.Errorln(err)
		}
		if err = a.triggerDeploymentRollout(agentCtx); err != nil {
			a.logger.Errorln(err)
		}

	} else {
		a.logger.Infoln("No configuration change detected for the agent")
	}
	if err := a.performAPMUpdation(ctx, &updateResp); err != nil {
		a.logger.Errorln(err)
	}
	return nil
}

func (a *K8sConfigUpdater) UpdateConfigMap(daemonSetConfig map[string]interface{}, deploymentConfig map[string]interface{}) error {
	daemonSetYamlBytes, err := yaml.Marshal(daemonSetConfig)
	if err != nil {
		return fmt.Errorf("marshal error for DaemonSet otel-config: %w", err)
	}

	deploymentYamlBytes, err := yaml.Marshal(deploymentConfig)
	if err != nil {
		return fmt.Errorf("marshal error for Deployment otel-config: %w", err)
	}

	configMaps := a.cfg.K8sClient.CoreV1().ConfigMaps(a.cfg.KubeNamespace)

	a.logger.Infoln("Attempting to update DaemonSet configMap.")
	_, err = configMaps.Update(context.TODO(), &corev1.ConfigMap{
		Data: map[string]string{"agent-daemonset.yaml": string(daemonSetYamlBytes)},
		ObjectMeta: v1.ObjectMeta{
			Name: a.cfg.ConfigmapDaemonsetName,
		},
	}, v1.UpdateOptions{})

	if err != nil {
		return fmt.Errorf("failed to update DaemonsSt configMap: %w", err)
	} else {
		a.logger.Infoln("Successfully updated DaemonSet configMap.")
	}

	a.logger.Infoln("Attempting to update Deployment configMap.")
	_, err = configMaps.Update(context.TODO(), &corev1.ConfigMap{
		Data: map[string]string{"agent-deployment.yaml": string(deploymentYamlBytes)},
		ObjectMeta: v1.ObjectMeta{
			Name: a.cfg.ConfigmapDeploymentName,
		},
	}, v1.UpdateOptions{})

	if err != nil {
		return fmt.Errorf("failed to update Deployment configMap: %w", err)
	} else {
		a.logger.Infoln("Successfully updated Deployment configMap.")
	}
	return nil
}

// triggerDaemonSetRollout triggers a DaemonSet rollout by patching its template annotation.
func (drt *K8sConfigUpdater) triggerDaemonSetRollout(ctx context.Context) error {
	drt.logger.Infof("Attempting to trigger rollout for DaemonSet %s/%s...", drt.cfg.KubeNamespace, drt.cfg.DaemonSetName)

	// Get the DaemonSet to ensure it exists and get its current state
	_, err := drt.cfg.K8sClient.AppsV1().DaemonSets(drt.cfg.KubeNamespace).Get(ctx, drt.cfg.DaemonSetName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Error getting DaemonSet %s/%s: %v", drt.cfg.KubeNamespace, drt.cfg.DaemonSetName, err)
	}

	// Prepare the patch to update the "kubectl.kubernetes.io/restartedAt" annotation.
	// This annotation change signals Kubernetes to perform a rolling update.
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("Error marshaling patch for DaemonSet %s/%s: %v", drt.cfg.KubeNamespace, drt.cfg.DaemonSetName, err)
	}

	// Apply the strategic merge patch to the DaemonSet
	_, err = drt.cfg.K8sClient.AppsV1().DaemonSets(drt.cfg.KubeNamespace).Patch(ctx, drt.cfg.DaemonSetName, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("Error patching DaemonSet %s/%s to trigger rollout: %v", drt.cfg.KubeNamespace, drt.cfg.DaemonSetName, err)
	}

	drt.logger.Infof("Successfully triggered rollout for DaemonSet %s/%s.", drt.cfg.KubeNamespace, drt.cfg.DaemonSetName)
	return nil
}

// triggerDeploymentRollout triggers a Deployment rollout by patching its template annotation.
func (drt *K8sConfigUpdater) triggerDeploymentRollout(ctx context.Context) error {
	drt.logger.Infof("Attempting to trigger rollout for Deployment %s/%s...", drt.cfg.KubeNamespace, drt.cfg.DeploymentName)

	// Get the Deployment to ensure it exists and get its current state
	_, err := drt.cfg.K8sClient.AppsV1().Deployments(drt.cfg.KubeNamespace).Get(ctx, drt.cfg.DeploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Error getting Deployment %s/%s: %v", drt.cfg.KubeNamespace, drt.cfg.DeploymentName, err)
	}

	// Prepare the patch to update the "kubectl.kubernetes.io/restartedAt" annotation.
	// This annotation change signals Kubernetes to perform a rolling update.
	patch := map[string]interface{}{
		"spec": map[string]interface{}{
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.RFC3339),
					},
				},
			},
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("Error marshaling patch for Deployment %s/%s: %v", drt.cfg.KubeNamespace, drt.cfg.DeploymentName, err)
	}

	// Apply the strategic merge patch to the Deployment
	_, err = drt.cfg.K8sClient.AppsV1().Deployments(drt.cfg.KubeNamespace).Patch(ctx, drt.cfg.DeploymentName, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("Error patching Deployment %s/%s to trigger rollout: %v", drt.cfg.KubeNamespace, drt.cfg.DeploymentName, err)
	}

	drt.logger.Infof("Successfully triggered rollout for Deployment %s/%s.", drt.cfg.KubeNamespace, drt.cfg.DeploymentName)
	return nil
}

func (a *K8sConfigUpdater) performAPMUpdation(ctx context.Context, response *K8sConfigUpdateResponse) error {

	if !response.K8s.APMEnabled {
		a.logger.Infof("Apm is not enabled for  %s\n", a.cfg.ClusterName)
		return nil
	}
	a.logger.Infof("Performing APM updation to cluster :%s on %d apps", a.cfg.ClusterName, len(response.K8s.APMSettings))
	for _, app := range response.K8s.APMSettings {
		kind := strings.ToUpper(app.Kind)
		annotations := instrumentation.KmCrdAnnotation(app.Language, app.Enabled)
		annotationBytes, err := json.Marshal(annotations)
		if err != nil {
			return fmt.Errorf("error marshaling patch %s/%s: %v", app.Namespace, app.Deployment, err)
		}
		// fetch existing resourse then check if annotation already exists else apply them
		switch kind {
		case "DAEMONSET":
			ds, err := a.cfg.K8sClient.AppsV1().DaemonSets(app.Namespace).Get(ctx, app.Deployment, v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}
			if isApplied := isAnnotationSame(annotations, ds.Spec.Template.Annotations); !isApplied {
				a.logger.Infof("[APM]: annotation for : %s using %s of kind : %s already applied \n", app.Deployment, app.Language, app.Kind)
				continue
			}
			_, err = a.cfg.K8sClient.AppsV1().DaemonSets(app.Namespace).Patch(ctx, app.Deployment, types.StrategicMergePatchType, annotationBytes, v1.PatchOptions{})
			if err != nil {
				return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}
		case "REPLICASET":
			rs, err := a.cfg.K8sClient.AppsV1().ReplicaSets(app.Namespace).Get(ctx, app.Deployment, v1.GetOptions{})
			// if err is not nil means replicaset doest not exist or has a parent deployment
			if err != nil {
				if errors.IsNotFound(err) {
					// if replicaset not found means replicaset has deployment has parent so apply patch on deployment
					dep, err := a.cfg.K8sClient.AppsV1().Deployments(app.Namespace).Get(ctx, app.Deployment, v1.GetOptions{})
					if err != nil {
						return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
					}
					if isApplied := isAnnotationSame(annotations, dep.Spec.Template.Annotations); !isApplied {
						a.logger.Infof("[APM]: annotation for : %s using %s of kind : %s already applied \n", app.Deployment, app.Language, app.Kind)
						continue
					} else {
						if err := handleDeploymentPatching(ctx, a.cfg.K8sClient, app, annotationBytes); err != nil {
							return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
						}
					}
				}
				return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			} else {
				// err is nil means replicaset exist and patch can be applied on it
				if isApplied := isAnnotationSame(annotations, rs.Spec.Template.Annotations); !isApplied {
					a.logger.Infof("[APM]: annotation for : %s using %s of kind : %s already applied \n", app.Deployment, app.Language, app.Kind)
					continue
				}
				_, err = a.cfg.K8sClient.AppsV1().ReplicaSets(app.Namespace).Patch(ctx, app.Deployment, types.StrategicMergePatchType, annotationBytes, v1.PatchOptions{})
				if err != nil {
					return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Deployment, app.Deployment, err)
				}
			}

		case "DEPLOYMENT":
			ds, err := a.cfg.K8sClient.AppsV1().Deployments(app.Namespace).Get(ctx, app.Deployment, v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}
			if isApplied := isAnnotationSame(annotations, ds.Spec.Template.Annotations); !isApplied {
				a.logger.Infof("[APM]: annotation for : %s using %s of kind : %s already applied \n", app.Deployment, app.Language, app.Kind)
				continue
			}
			if err := handleDeploymentPatching(ctx, a.cfg.K8sClient, app, annotationBytes); err != nil {
				return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}
		case "STATEFULSET":
			ss, err := a.cfg.K8sClient.AppsV1().StatefulSets(app.Namespace).Get(ctx, app.Deployment, v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}
			if isApplied := isAnnotationSame(annotations, ss.Spec.Template.Annotations); !isApplied {
				a.logger.Infof("[APM]: annotation for : %s using %s of kind : %s already applied \n", app.Deployment, app.Language, app.Kind)
				continue
			}
			_, err = a.cfg.K8sClient.AppsV1().StatefulSets(app.Namespace).Patch(ctx, app.Deployment, types.StrategicMergePatchType, annotationBytes, v1.PatchOptions{})
			if err != nil {
				return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}

		case "POD":
			pod, err := a.cfg.K8sClient.CoreV1().Pods(app.Namespace).Get(ctx, app.Deployment, v1.GetOptions{})
			if err != nil {
				return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}
			if isApplied := isAnnotationSame(annotations, pod.Annotations); !isApplied {
				a.logger.Infof("[APM]: annotation for : %s using %s of kind : %s already applied \n", app.Deployment, app.Language, app.Kind)
				continue
			}
			_, err = a.cfg.K8sClient.CoreV1().Pods(app.Namespace).Patch(ctx, app.Deployment, types.StrategicMergePatchType, annotationBytes, v1.PatchOptions{})
			if err != nil {
				return fmt.Errorf("error applying auto instrumentation on %s/%s: %v", app.Namespace, app.Deployment, err)
			}
		default:
			return fmt.Errorf("error applying auto instrumentation invalid KIND provided %s", app.Kind)
		}
	}

	return nil
}

func isAnnotationSame(annotations instrumentation.InstrumentAnnotiation, resourceMap map[string]string) bool {
	for key, value := range annotations {
		// Look up the key in the second map.
		if val2, ok := resourceMap[key]; !ok || val2 != value {
			// If the key does not exist (!ok) or the value is different (val2 != value),
			// then map1 is not a subset of map2. Return false immediately.
			return false
		}
	}
	return true
}

func handleDeploymentPatching(ctx context.Context, client *kubernetes.Clientset, app APMConfig, annotations []byte) error {
	_, err := client.AppsV1().Deployments(app.Namespace).Patch(ctx, app.Deployment, types.StrategicMergePatchType, annotations, v1.PatchOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *K8sConfigUpdater) otelConfigPath() string {
	daemonsetURI := "/etc/kmagent/agent-daemonset.yaml"
	deploymentURI := "/etc/kmagent/agent-deployment.yaml"
	if c.cfg.DeploymentMode == "DAEMONSET" {
		return daemonsetURI
	} else {
		return deploymentURI
	}
}

func (c *K8sConfigUpdater) SetConfigPath() {
	c.configPath = c.otelConfigPath()
}
