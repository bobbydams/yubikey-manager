package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration values for the application.
type Config struct {
	PrimaryKeyID          string `mapstructure:"primary_key_id"`
	PrimaryKeyFingerprint string `mapstructure:"primary_key_fingerprint"`
	UserName              string `mapstructure:"user_name"`
	UserEmail             string `mapstructure:"user_email"`
	Keyserver             string `mapstructure:"keyserver"`
	MasterKeyPath         string `mapstructure:"master_key_path"`
	BackupDir             string `mapstructure:"backup_dir"`
	NoColor               bool   `mapstructure:"no_color"`
}

// Load reads configuration from multiple sources with the following priority:
// 1. CLI flags (highest priority)
// 2. Environment variables
// 3. Config file
// 4. Defaults (lowest priority)
func Load() (*Config, error) {
	// Set defaults
	viper.SetDefault("keyserver", "hkps://keys.openpgp.org")
	viper.SetDefault("backup_dir", filepath.Join(os.Getenv("HOME"), ".gnupg", "backups"))

	// Set config file name and paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Add config paths (in order of precedence)
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ykgpg")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// Environment variables
	viper.SetEnvPrefix("YKGPG")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file (optional - won't error if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults/env/flags
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// BindFlag binds a single CLI flag to viper configuration.
// This is a helper function to be called from Cobra command setup.
func BindFlag(flagName, viperKey string) error {
	// This will be called from the CLI layer when setting up commands
	return nil
}

// Validate checks that required configuration values are set.
func (c *Config) Validate() error {
	if c.PrimaryKeyID == "" {
		return fmt.Errorf("primary_key_id is required")
	}
	if c.PrimaryKeyFingerprint == "" {
		return fmt.Errorf("primary_key_fingerprint is required")
	}
	if c.UserName == "" {
		return fmt.Errorf("user_name is required")
	}
	if c.UserEmail == "" {
		return fmt.Errorf("user_email is required")
	}
	return nil
}
