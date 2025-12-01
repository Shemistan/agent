package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
)

// Config represents the application configuration
type Config struct {
	ServiceName string      `toml:"service_name"`
	ServiceEnv  string      `toml:"service_env"`
	HTTPPort    int         `toml:"http_port"`
	Database    DatabaseCfg `toml:"database"`
	Manager     ManagerCfg  `toml:"manager"`
}

// DatabaseCfg represents database configuration
type DatabaseCfg struct {
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
	Name    string `toml:"name"`
	SSLMode string `toml:"sslmode"`
}

// ManagerCfg represents manager service configuration
type ManagerCfg struct {
	URL            string `toml:"url"`
	TimeoutSeconds int    `toml:"timeout_seconds"`
}

// Load loads configuration from app.toml and environment variables
func Load(configPath string) (*Config, error) {
	// Load .env file if in local environment
	_ = godotenv.Load(".env")

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Override database credentials from environment
	if user := os.Getenv("DB_USER"); user != "" {
		// Will be used later in DSN
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		// Will be used later in DSN
	}

	return &cfg, nil
}

// GetDSN returns PostgreSQL DSN string
func (c *Config) GetDSN() string {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")

	if user == "" {
		user = "postgres"
	}
	if password == "" {
		password = "postgres"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, c.Database.Host, c.Database.Port, c.Database.Name, c.Database.SSLMode,
	)
}

// GetManagerURL returns the manager service URL
func (c *Config) GetManagerURL() string {
	return c.Manager.URL
}

// GetManagerTimeout returns the manager timeout in seconds
func (c *Config) GetManagerTimeout() int {
	return c.Manager.TimeoutSeconds
}
