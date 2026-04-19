package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

// Config holds the application configuration.
type Config struct {
	VaultAddr  string `mapstructure:"vault_addr"`
	VaultToken string `mapstructure:"vault_token"`
	SecretPath string `mapstructure:"secret_path"`
	OutputFile string `mapstructure:"output_file"`
	MountPath  string `mapstructure:"mount_path"`
}

// Load reads configuration from a file and environment variables.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	v.SetDefault("vault_addr", "http://127.0.0.1:8200")
	v.SetDefault("output_file", ".env")
	v.SetDefault("mount_path", "secret")

	v.SetEnvPrefix("VAULTPULL")
	v.AutomaticEnv()

	// Also respect native Vault env vars.
	if addr := os.Getenv("VAULT_ADDR"); addr != "" {
		v.Set("vault_addr", addr)
	}
	if token := os.Getenv("VAULT_TOKEN"); token != "" {
		v.Set("vault_token", token)
	}

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName(".vaultpull")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME")
	}

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.VaultToken == "" {
		return nil, errors.New("vault token is required (set VAULT_TOKEN or vault_token in config)")
	}
	if cfg.SecretPath == "" {
		return nil, errors.New("secret_path is required")
	}

	return &cfg, nil
}
