package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
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

func loadRemoteConfig(url string) error {
	resp, err := http.Get(url)
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
