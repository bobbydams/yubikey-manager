package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bobbydams/yubikey-manager/internal/config"
	"github.com/bobbydams/yubikey-manager/pkg/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long:  "Commands for managing YubiKey GPG Manager configuration",
	}
	// Skip PersistentPreRunE validation for config commands
	// These commands need to work even without a valid config file
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Do nothing - config commands handle their own validation
		return nil
	}

	cmd.AddCommand(newConfigInitCmd())
	cmd.AddCommand(newConfigShowCmd())

	return cmd
}

func newConfigInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Interactively generate configuration file",
		Long: `Interactively generate a configuration file at ~/.config/ykgpg/config.yaml.
This command will prompt you for all required configuration values.`,
		RunE: runConfigInit,
	}
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration values",
		Long: `Display the current configuration values from all sources:
- CLI flags (highest priority)
- Environment variables
- Config file
- Defaults (lowest priority)`,
		RunE: runConfigShow,
	}
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	ui.PrintHeader("Generate Configuration File")

	fmt.Println("This will create a configuration file at ~/.config/ykgpg/config.yaml")
	fmt.Println("You can override these values later with environment variables or CLI flags.")
	fmt.Println()

	cfg := &config.Config{}

	// Prompt for required values
	var err error

	cfg.PrimaryKeyID, err = ui.PromptRequired("Primary Key ID (e.g., ABC123DEF4567890): ")
	if err != nil {
		return err
	}

	cfg.PrimaryKeyFingerprint, err = ui.PromptRequired("Primary Key Fingerprint (full 40-char hex): ")
	if err != nil {
		return err
	}

	cfg.UserName, err = ui.PromptRequired("Your Name: ")
	if err != nil {
		return err
	}

	cfg.UserEmail, err = ui.PromptRequired("Your Email: ")
	if err != nil {
		return err
	}

	// Optional values with defaults
	keyserver, err := ui.Prompt("Keyserver URL [hkps://keys.openpgp.org]: ")
	if err != nil {
		return err
	}
	if keyserver == "" {
		cfg.Keyserver = "hkps://keys.openpgp.org"
	} else {
		cfg.Keyserver = keyserver
	}

	backupDir, err := ui.Prompt(fmt.Sprintf("Backup directory [%s/.gnupg/backups]: ", os.Getenv("HOME")))
	if err != nil {
		return err
	}
	if backupDir == "" {
		cfg.BackupDir = filepath.Join(os.Getenv("HOME"), ".gnupg", "backups")
	} else {
		cfg.BackupDir = backupDir
	}

	masterKeyPath, err := ui.Prompt("Master key path (optional, can be set later): ")
	if err != nil {
		return err
	}
	cfg.MasterKeyPath = masterKeyPath

	noColorStr, err := ui.Prompt("Disable colored output? [N]: ")
	if err != nil {
		return err
	}
	cfg.NoColor = (noColorStr == "y" || noColorStr == "Y" || noColorStr == "yes" || noColorStr == "Yes")

	// Validate the config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create config directory
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ykgpg")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	configFile := filepath.Join(configDir, "config.yaml")
	configData := map[string]interface{}{
		"primary_key_id":          cfg.PrimaryKeyID,
		"primary_key_fingerprint": cfg.PrimaryKeyFingerprint,
		"user_name":               cfg.UserName,
		"user_email":              cfg.UserEmail,
		"keyserver":               cfg.Keyserver,
		"backup_dir":              cfg.BackupDir,
		"no_color":                cfg.NoColor,
	}
	if cfg.MasterKeyPath != "" {
		configData["master_key_path"] = cfg.MasterKeyPath
	}

	yamlData, err := yaml.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	ui.LogSuccess("Configuration file created at: %s", configFile)
	fmt.Println()
	fmt.Println("You can now use ykgpg commands. Configuration can be overridden with:")
	fmt.Println("  - Environment variables (YKGPG_* prefix)")
	fmt.Println("  - CLI flags (--key-id, --name, etc.)")
	fmt.Println()

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Load config (this will use the normal priority: flags > env > file > defaults)
	cfg, err := config.Load()
	if err != nil {
		// If config loading fails, show what we can
		ui.LogWarning("Failed to load full config: %v", err)
		fmt.Println()
		fmt.Println("Showing available configuration sources:")
		fmt.Println()

		// Show environment variables
		fmt.Println("Environment Variables:")
		fmt.Println("  YKGPG_PRIMARY_KEY_ID:", os.Getenv("YKGPG_PRIMARY_KEY_ID"))
		fmt.Println("  YKGPG_PRIMARY_KEY_FINGERPRINT:", os.Getenv("YKGPG_PRIMARY_KEY_FINGERPRINT"))
		fmt.Println("  YKGPG_USER_NAME:", os.Getenv("YKGPG_USER_NAME"))
		fmt.Println("  YKGPG_USER_EMAIL:", os.Getenv("YKGPG_USER_EMAIL"))
		fmt.Println("  YKGPG_KEYSERVER:", os.Getenv("YKGPG_KEYSERVER"))
		fmt.Println("  YKGPG_MASTER_KEY_PATH:", os.Getenv("YKGPG_MASTER_KEY_PATH"))
		fmt.Println("  YKGPG_BACKUP_DIR:", os.Getenv("YKGPG_BACKUP_DIR"))
		fmt.Println()

		// Show config file location
		configFile := filepath.Join(os.Getenv("HOME"), ".config", "ykgpg", "config.yaml")
		if _, err := os.Stat(configFile); err == nil {
			fmt.Printf("Config file exists: %s\n", configFile)
		} else {
			fmt.Printf("Config file not found: %s\n", configFile)
		}

		return err
	}

	ui.PrintHeader("Current Configuration")

	fmt.Println("Configuration values (showing effective values from all sources):")
	fmt.Println()
	ui.PrintKeyValueKey("Primary Key ID", cfg.PrimaryKeyID)
	ui.PrintKeyValueKey("Primary Key Fingerprint", cfg.PrimaryKeyFingerprint)
	ui.PrintKeyValue("User Name", cfg.UserName)
	ui.PrintKeyValue("User Email", cfg.UserEmail)
	ui.PrintKeyValue("Keyserver", cfg.Keyserver)
	if cfg.MasterKeyPath != "" {
		ui.PrintKeyValue("Master Key Path", cfg.MasterKeyPath)
	} else {
		ui.PrintKeyValue("Master Key Path", "(not set)")
	}
	ui.PrintKeyValue("Backup Directory", cfg.BackupDir)
	fmt.Println()

	// Show where values come from
	fmt.Println("Configuration Sources:")
	configFile := filepath.Join(os.Getenv("HOME"), ".config", "ykgpg", "config.yaml")
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("  ✓ Config file: %s\n", configFile)
	} else {
		fmt.Printf("  ✗ Config file: %s (not found)\n", configFile)
	}

	// Check for environment variables
	envVars := []string{
		"YKGPG_PRIMARY_KEY_ID",
		"YKGPG_PRIMARY_KEY_FINGERPRINT",
		"YKGPG_USER_NAME",
		"YKGPG_USER_EMAIL",
		"YKGPG_KEYSERVER",
		"YKGPG_MASTER_KEY_PATH",
		"YKGPG_BACKUP_DIR",
	}
	hasEnvVars := false
	for _, envVar := range envVars {
		if os.Getenv(envVar) != "" {
			if !hasEnvVars {
				fmt.Println("  ✓ Environment variables: (some set)")
				hasEnvVars = true
			}
			break
		}
	}
	if !hasEnvVars {
		fmt.Println("  ✗ Environment variables: (none set)")
	}

	// Check for CLI flags
	fmt.Println("  ℹ CLI flags: (check with --help for available flags)")
	fmt.Println()

	return nil
}
