//go:build k8s
// +build k8s

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
	"time"

	"github.com/kloudmate/km-agent/internal/config"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// ConfigUpdater handles configuration updates from a remote API
type K8sConfigUpdater struct {
	cfg        *config.K8sAgentConfig
	logger     *zap.SugaredLogger
	client     *http.Client
	configPath string
}

type K8sUpdateCheckerParams struct {
	Version            string
	AgentStatus        string
	CollectorStatus    string
	CollectorLastError string
}

// ConfigUpdateResponse represents the response from the config update API
type K8sConfigUpdateResponse struct {
	DeploymentType string                 `json:"deployment_type"`
	Config         map[string]interface{} `json:"config"`
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
func (u *K8sConfigUpdater) CheckForUpdates(ctx context.Context, p K8sUpdateCheckerParams) (string, map[string]interface{}, error) {

	// Create the request
	data := map[string]interface{}{
		"is_docker":          false,
		"platform":           "kubernetes",
		"architecture":       runtime.GOARCH,
		"agent_version":      p.Version,
		"agent_status":       p.AgentStatus,
		"collector_status":   p.CollectorStatus,
		"last_error_message": p.CollectorLastError,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel() // Ensure context resources are freed

	req, err := http.NewRequestWithContext(reqCtx, "POST", u.cfg.ConfigUpdateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Add API key if configured
	if u.cfg.APIKey != "" {
		req.Header.Set("Authorization", u.cfg.APIKey)
	}

	resp, respErr := u.client.Do(req)

	if respErr != nil {
		return "", nil, fmt.Errorf("failed to fetch config updates after retries: %w", respErr)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("config update API returned non-OK status: %d, body: %s", resp.StatusCode, body)
	}

	// Parse response
	var updateResp K8sConfigUpdateResponse
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		return "", nil, fmt.Errorf("failed to decode config update response: %w", err)
	}

	return updateResp.DeploymentType, updateResp.Config, nil
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
	ctx, cancel := context.WithTimeout(agentCtx, 10*time.Second)
	defer cancel()

	a.logger.Infoln("Checking for configuration updates...")

	params := K8sUpdateCheckerParams{
		Version: a.cfg.Version,
	}

	a.logger.Debugf("Checking for updates with params: %+v", params)

	deploymentMode, newConfig, err := a.CheckForUpdates(ctx, params)
	if err != nil {
		return fmt.Errorf("updater.CheckForUpdates failed: %w", err)
	}
	if newConfig != nil && deploymentMode != "" {
		if err := a.UpdateConfigMap(newConfig, deploymentMode); err != nil {
			return fmt.Errorf("failed to update config file: %w", err)
		}
		a.logger.Infoln("triggering rollout restart.")

		if err = a.triggerRollout(agentCtx); err != nil {
			a.logger.Errorln(err)
		}

	} else {
		a.logger.Infoln("No configuration change")
	}
	return nil
}

func (a *K8sConfigUpdater) UpdateConfigMap(cfg map[string]interface{}, deploymentMode string) error {

	yamlBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	configMaps := a.cfg.K8sClient.CoreV1().ConfigMaps(a.cfg.KubeNamespace)

	if deploymentMode == "DAEMONSET" {
		_, err = configMaps.Update(context.TODO(), &corev1.ConfigMap{
			Data: map[string]string{"agent-daemonset.yaml": string(yamlBytes)},
			ObjectMeta: v1.ObjectMeta{
				Name: a.cfg.ConfigmapDaemonsetName,
			},
		}, v1.UpdateOptions{})
	} else {
		_, err = configMaps.Update(context.TODO(), &corev1.ConfigMap{
			Data: map[string]string{"agent-deployment.yaml": string(yamlBytes)},
			ObjectMeta: v1.ObjectMeta{
				Name: a.cfg.ConfigmapDeploymentName,
			},
		}, v1.UpdateOptions{})
	}

	if err != nil {
		a.logger.Errorln(err)
	}

	return nil
}

// triggerRollout triggers a DaemonSet rollout by patching its template annotation.
func (drt *K8sConfigUpdater) triggerRollout(ctx context.Context) error {
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
