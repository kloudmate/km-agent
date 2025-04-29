package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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

// ConfigUpdateResponse represents the response from the config update API
type ConfigUpdateResponse struct {
	RestartRequired bool                   `json:"restart_required"`
	Config          map[string]interface{} `json:"config"`
}

// NewConfigUpdater creates a new config updater
func NewConfigUpdater(cfg *config.Config, logger *zap.SugaredLogger) *ConfigUpdater {
	// Determine config path
	configPath := cfg.ConfigPath
	if configPath == "" {
		if cfg.Agent.DockerMode {
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
func (u *ConfigUpdater) CheckForUpdates(ctx context.Context) (bool, map[string]interface{}, error) {
	// If no config update URL is configured, return nil
	if u.cfg.Agent.ConfigUpdateURL == "" {
		return false, nil, nil
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "GET", u.cfg.Agent.ConfigUpdateURL, nil)
	if err != nil {
		return false, nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add API key if configured
	if u.cfg.Agent.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+u.cfg.Agent.APIKey)
	}

	// Execute the request with retry logic
	var resp *http.Response
	var respErr error

	// Retry 3 times with exponential backoff
	for attempt := 0; attempt < 3; attempt++ {
		resp, respErr = u.client.Do(req)
		if respErr == nil {
			break
		}

		// Exponential backoff: 1s, 2s, 4s
		backoffTime := time.Duration(1<<attempt) * time.Second
		u.logger.Warnf("Request failed (attempt %d/3): %v, retrying in %v", attempt+1, respErr, backoffTime)

		select {
		case <-time.After(backoffTime):
			// Continue with retry
		case <-ctx.Done():
			return false, nil, ctx.Err()
		}
	}

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

	// If there's a new config, write it to disk
	if updateResp.Config != nil {
		if err := u.ApplyConfig(updateResp.Config); err != nil {
			return false, nil, fmt.Errorf("failed to apply new config: %w", err)
		}
	}

	// Return the update info
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
	return nil
}
