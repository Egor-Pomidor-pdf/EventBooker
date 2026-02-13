package config

type PostgresConfig struct {
	MasterDSN                    string   `env:"MASTER_DSN"`
	SlaveDSNs                    []string `env:"SLAVE_DSNS" envSeparator:","`
	MaxOpenConnections           int      `env:"MAX_OPEN_CONNECTIONS" envDefault:"3"`
	MaxIdleConnections           int      `env:"MAX_IDLE_CONNECTIONS" envDefault:"5"`
	ConnectionMaxLifetimeSeconds int      `env:"CONNECTION_MAX_LIFETIME_SECONDS" envDefault:"0"`
	PostgresRetryConfig RetryConfig 
}

type ServerConfig struct {
	Host string `env:"SERVER_HOST"` 
	Port int    `yaml:"SERVER_PORT"` 
}

type RetryConfig struct {
	Attempts          int     `env:"ATTEMPTS"`
	DelayMilliseconds int     `env:"DELAY_MS"`
	Backoff           float64 `env:"BACKOFF"`
}

