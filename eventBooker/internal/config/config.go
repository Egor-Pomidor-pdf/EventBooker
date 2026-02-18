package config

import (
	"fmt"
	"time"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/retry"
)

type Config struct {
	Env      string `yaml:"env" env:"ENV"`
	Server   ServerConfig
	Database PostgresConfig
}

func NewConfig(envFilePath string, configFilePath string) (*Config, error) {
	myConfig := &Config{}

	cfg := config.New()

	if envFilePath != "" {
		if err := cfg.LoadEnvFiles(envFilePath); err != nil {
			return nil, fmt.Errorf("failed to load .env file: %w", err)
		}
	}
	cfg.EnableEnv("")

	if configFilePath != "" {
		if err := cfg.LoadConfigFiles(configFilePath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	myConfig.Env = cfg.GetString("ENV")
	// Postgres
	myConfig.Database.MasterDSN = cfg.GetString("POSTGRES_MASTER_DSN")
	myConfig.Database.SlaveDSNs = cfg.GetStringSlice("POSTGRES_SLAVE_DSNS")
	myConfig.Database.MaxOpenConnections = cfg.GetInt("POSTGRES_MAX_OPEN_CONNECTIONS")
	myConfig.Database.MaxIdleConnections = cfg.GetInt("POSTGRES_MAX_IDLE_CONNECTIONS")
	myConfig.Database.ConnectionMaxLifetimeSeconds = cfg.GetInt("POSTGRES_CONNECTION_MAX_LIFETIME_SECONDS")
	// Postgres retry
	myConfig.Database.PostgresRetryConfig.Attempts = cfg.GetInt("RETRY_POSTGRES_ATTEMPTS")
	myConfig.Database.PostgresRetryConfig.DelayMilliseconds = cfg.GetInt("RETRY_POSTGRES_DELAY_MS")
	myConfig.Database.PostgresRetryConfig.Backoff = cfg.GetFloat64("RETRY_POSTGRES_BACKOFF")
	//server
	myConfig.Server.Host = cfg.GetString("SERVER_HOST")
	myConfig.Server.Port = cfg.GetInt("SERVER_PORT")
	myConfig.Server.PprofAddr = cfg.GetString("PPROF_ADDR")

	return myConfig, nil
}

func MakeStrategy(c RetryConfig) retry.Strategy {
	return retry.Strategy{
		Attempts: c.Attempts,
		Delay:    time.Duration(c.DelayMilliseconds) * time.Millisecond,
		Backoff:  c.Backoff,
	}
}
