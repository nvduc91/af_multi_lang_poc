package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	AppName     string            `json:"app_name" mapstructure:"app_name"`
	Version     string            `json:"version" mapstructure:"version"`
	Port        int               `json:"port" mapstructure:"port"`
	Environment string            `json:"environment" mapstructure:"environment"`
	Database    DatabaseConfig    `json:"database" mapstructure:"database"`
	Features    map[string]bool   `json:"features" mapstructure:"features"`
	Metadata    map[string]string `json:"metadata" mapstructure:"metadata"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host     string `json:"host" mapstructure:"host"`
	Port     int    `json:"port" mapstructure:"port"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
	Database string `json:"database" mapstructure:"database"`
}

func loadConfig(configType string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/app/config")

	// Set default values
	viper.SetDefault("app_name", "go-sample-app")
	viper.SetDefault("version", "1.0.0")
	viper.SetDefault("port", 8080)
	viper.SetDefault("environment", "development")

	// Load config based on type
	switch configType {
	case "local":
		// Read local config file
		if err := viper.ReadInConfig(); err != nil {
			logrus.Warnf("Failed to read local config file: %v", err)
		}
	case "remote":
		// Load from remote URL
		if remoteURL := os.Getenv("CONFIG_REMOTE_URL"); remoteURL != "" {
			if err := loadRemoteConfig(remoteURL); err != nil {
				logrus.Warnf("Failed to load remote config: %v", err)
			}
		}
	case "env":
		// Only use environment variables
		logrus.Info("Using environment variables for configuration")
	default:
		// Try local first, then remote
		if err := viper.ReadInConfig(); err != nil {
			logrus.Warnf("Failed to read local config file: %v", err)
		}
		if remoteURL := os.Getenv("CONFIG_REMOTE_URL"); remoteURL != "" {
			if err := loadRemoteConfig(remoteURL); err != nil {
				logrus.Warnf("Failed to load remote config: %v", err)
			}
		}
	}

	// Bind environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}

func loadRemoteConfig(configURL string) error {
	// Convert GitHub web URL to raw content URL if needed
	rawURL := convertToRawURL(configURL)
	logrus.Infof("Loading remote config from: %s", rawURL)

	// Create HTTP client with authentication for private repos
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add GitHub token if available for private repos
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "token "+token)
		logrus.Info("Using GitHub token for authentication")
	}

	// Add User-Agent header (GitHub requires this)
	req.Header.Set("User-Agent", "go-sample-app/1.0")

	// Add cache-busting headers to ensure we get the latest version
	req.Header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Expires", "0")

	// Add timestamp to URL to bypass cache
	timestamp := time.Now().Unix()
	cacheBustURL := rawURL
	if strings.Contains(rawURL, "?") {
		cacheBustURL += fmt.Sprintf("&_t=%d", timestamp)
	} else {
		cacheBustURL += fmt.Sprintf("?_t=%d", timestamp)
	}
	req.URL, _ = url.Parse(cacheBustURL)

	logrus.Infof("Making request to: %s", req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch remote config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch remote config, status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	logrus.Infof("Received config content (%d bytes):", len(body))
	logrus.Infof("Config content preview: %s", string(body[:min(len(body), 200)]))

	if err := viper.ReadConfig(bytes.NewReader(body)); err != nil {
		// Try to parse as JSON
		var jsonConfig map[string]interface{}
		if err := json.Unmarshal(body, &jsonConfig); err != nil {
			return fmt.Errorf("failed to parse remote config as YAML or JSON: %w", err)
		}
		for key, value := range jsonConfig {
			viper.Set(key, value)
		}
	}
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// convertToRawURL converts GitHub web URLs to raw content URLs
func convertToRawURL(url string) string {
	// If it's already a raw URL, return as is
	if len(url) > 7 && url[:7] == "https://raw.githubusercontent.com" {
		return url
	}

	// Convert GitHub web URL to raw URL
	// From: https://github.com/user/repo/blob/branch/path/to/file
	// To:   https://raw.githubusercontent.com/user/repo/branch/path/to/file
	if len(url) > 18 && url[:18] == "https://github.com" {
		// Replace "github.com" with "raw.githubusercontent.com"
		// Remove "/blob" from the path
		url = "https://raw.githubusercontent.com" + url[18:]
		url = strings.Replace(url, "/blob/", "/", 1)
		return url
	}

	// Support SSH URLs (for private repos with SSH access)
	// From: git@github.com:user/repo.git/blob/branch/path/to/file
	// To:   https://raw.githubusercontent.com/user/repo/branch/path/to/file
	if strings.HasPrefix(url, "git@github.com:") {
		// Extract the path after github.com:
		path := strings.TrimPrefix(url, "git@github.com:")
		// Remove .git if present
		path = strings.TrimSuffix(path, ".git")
		// Remove /blob/ and convert to raw URL
		path = strings.Replace(path, "/blob/", "/", 1)
		return "https://raw.githubusercontent.com/" + path
	}

	return url
}

func main() {
	// Parse command line flags
	configType := flag.String("config-type", "default", "Configuration type: local, env, remote, or default")
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.Infof("Starting go-sample-app with config type: %s", *configType)
	start := time.Now()

	config, err := loadConfig(*configType)
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	logrus.Infof("Configuration loaded in %s", time.Since(start))
	logrus.Info("Final configuration (JSON):")
	jsonBytes, _ := json.MarshalIndent(config, "", "  ")
	fmt.Println(string(jsonBytes))

	logrus.Info("Task completed successfully.")
}
