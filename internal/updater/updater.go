package updater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/kloudmate/km-agent/internal/config"
)

// ConfigUpdater handles configuration updates from a remote API
type ConfigUpdater struct {
	cfg        *config.Config
	logger     *zap.SugaredLogger
	client     *http.Client
	configPath string
}

type UpdateCheckerParams struct {
	Version            string
	AgentStatus        string
	CollectorStatus    string
	CollectorLastError string
}

// ConfigUpdateResponse represents the response from the config update API
type ConfigUpdateResponse struct {
	RestartRequired bool                   `json:"restart_required"`
	Config          map[string]interface{} `json:"config"`
}

// NewConfigUpdater creates a new config updater
func NewConfigUpdater(cfg *config.Config, logger *zap.SugaredLogger) *ConfigUpdater {
	// Determine config path
	configPath := cfg.OtelConfigPath
	if configPath == "" {
		if cfg.DockerMode {
			configPath = config.GetDockerConfigPath()
		} else {
			configPath = config.GetDefaultConfigPath()
		}
	}

	return &ConfigUpdater{
		cfg:        cfg,
		logger:     logger,
		client:     &http.Client{Timeout: 30 * time.Second},
		configPath: configPath,
	}
}

// CheckForUpdates checks for configuration updates from the remote API
func (u *ConfigUpdater) CheckForUpdates(ctx context.Context, p UpdateCheckerParams) (bool, map[string]interface{}, error) {

	// Create the request
	data := map[string]interface{}{
		"is_docker":          u.cfg.DockerMode,
		"hostname":           u.cfg.Hostname(),
		"platform":           runtime.GOOS,
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
func (u *ConfigUpdater) ApplyConfig(newConfig map[string]interface{}) error {
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
