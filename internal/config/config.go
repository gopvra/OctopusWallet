package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig            `mapstructure:"server"`
	Database DatabaseConfig          `mapstructure:"database"`
	Wallet   WalletConfig            `mapstructure:"wallet"`
	Chains   map[string]ChainConfig  `mapstructure:"chains"`
	Webhook  WebhookConfig           `mapstructure:"webhook"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	URL          string `mapstructure:"url"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type WalletConfig struct {
	MasterSeed    string `mapstructure:"master_seed"`
	EncryptionKey string `mapstructure:"encryption_key"`
}

type ChainConfig struct {
	Enabled               bool          `mapstructure:"enabled"`
	RPCURL                string        `mapstructure:"rpc_url"`
	ChainID               int64         `mapstructure:"chain_id"`
	ConfirmationsRequired uint64        `mapstructure:"confirmations_required"`
	PollInterval          time.Duration `mapstructure:"poll_interval"`
	APIKey                string        `mapstructure:"api_key"`
	RPCUser               string        `mapstructure:"rpc_user"`
	RPCPass               string        `mapstructure:"rpc_pass"`
	Network               string        `mapstructure:"network"`
}

type WebhookConfig struct {
	MaxRetries   int           `mapstructure:"max_retries"`
	RetryBackoff time.Duration `mapstructure:"retry_backoff"`
	Timeout      time.Duration `mapstructure:"timeout"`
}

func Load(path string) (*Config, error) {
	v := viper.New()

	// Defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("webhook.max_retries", 5)
	v.SetDefault("webhook.retry_backoff", "30s")
	v.SetDefault("webhook.timeout", "10s")

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	v.SetEnvPrefix("OCTOPUS")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
