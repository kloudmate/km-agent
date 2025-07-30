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
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kloudmate/km-agent/internal/config"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
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
	RestartRequired bool                   `json:"restart_required"`
	Config          map[string]interface{} `json:"config"`
}

// NewK8sConfigUpdater creates a new config updater
func NewK8sConfigUpdater(cfg *config.K8sAgentConfig, logger *zap.SugaredLogger) *K8sConfigUpdater {
	// Determine config path
	configPath := config.DefaultAgentConfigPath

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
		configPath: configPath,
	}
}

// CheckForUpdates checks for configuration updates from the remote API
func (u *K8sConfigUpdater) CheckForUpdates(ctx context.Context, p UpdateCheckerParams) (bool, map[string]interface{}, error) {

	// Create the request
	data := map[string]interface{}{
		"is_docker":          false,
		"hostname":           u.cfg.Hostname(),
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
		return false, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Add API key if configured
	if u.cfg.APIKey != "" {
		req.Header.Set("Authorization", u.cfg.APIKey)
	}

	resp, respErr := u.client.Do(req)

	if respErr != nil {
		return false, nil, fmt.Errorf("failed to fetch config updates after retries: %w", respErr)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, nil, fmt.Errorf("config update API returned non-OK status: %d, body: %s", resp.StatusCode, body)
	}

	// Parse response
	var updateResp ConfigUpdateResponse
	if err := json.NewDecoder(resp.Body).Decode(&updateResp); err != nil {
		return false, nil, fmt.Errorf("failed to decode config update response: %w", err)
	}

	return updateResp.RestartRequired, updateResp.Config, nil
}

// ApplyConfig applies a new configuration by writing it to the config file
func (u *K8sConfigUpdater) ApplyConfig(newConfig map[string]interface{}) error {
	// Convert to YAML
	configYAML, err := yaml.Marshal(newConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal new config to YAML: %w", err)
	}

	// Make sure the directory exists
	configDir := filepath.Dir(u.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create a temporary file in the same directory
	tempFile := u.configPath + ".new"
	if err := os.WriteFile(tempFile, configYAML, 0644); err != nil {
		return fmt.Errorf("failed to write new config to temporary file: %w", err)
	}

	// Rename the temporary file to the actual config file (atomic operation)
	if err := os.Rename(tempFile, u.configPath); err != nil {
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	u.logger.Info("Successfully updated configuration at ", u.configPath)

	defer func() {
		if err != nil {
			// Clean up temp file on error
			if removeErr := os.Remove(tempFile); removeErr != nil {
				u.logger.Warnf("Failed to clean up temporary file %s: %v", tempFile, removeErr)
			}
		}
	}()
	return nil
}
