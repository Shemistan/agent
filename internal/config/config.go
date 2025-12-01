package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
)

// DatabaseCfg represents database configuration
type DatabaseCfg struct {
	Host    string `toml:"host"`
	Port    int    `toml:"port"`
	User    string `toml:"user"`
	Name    string `toml:"name"`
	SSLMode string `toml:"sslmode"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled  bool   `toml:"enabled"`
	CertFile string `toml:"cert_file"`
	KeyFile  string `toml:"key_file"`
	CAFile   string `toml:"ca_file"`
}

// ManagerCfg represents manager service configuration
type ManagerCfg struct {
	URLs           []string `toml:"urls"`
	TimeoutSeconds int      `toml:"timeout_seconds"`
}

// Config represents the application configuration
type Config struct {
	ServiceName string      `toml:"service_name"`
	ServiceEnv  string      `toml:"service_env"`
	HTTPPort    int         `toml:"http_port"`
	Database    DatabaseCfg `toml:"database"`
	TLS         TLSConfig   `toml:"tls"`
	Manager     ManagerCfg  `toml:"manager"`
}

// Load loads configuration from app.toml and environment variables
func Load(configPath string) (*Config, error) {
	// Load .env file if in local environment
	_ = godotenv.Load(".env")

	var cfg Config
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	// Override with environment variables
	// Database configuration
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Database.Port)
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.Name = name
	}
	if sslmode := os.Getenv("DB_SSLMODE"); sslmode != "" {
		cfg.Database.SSLMode = sslmode
	}

	// App configuration
	if port := os.Getenv("APP_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.HTTPPort)
	}

	// TLS configuration
	if tlsEnabled := os.Getenv("TLS_ENABLED"); tlsEnabled != "" {
		cfg.TLS.Enabled = strings.ToLower(tlsEnabled) == "true"
	}
	if certFile := os.Getenv("TLS_CERT_FILE"); certFile != "" {
		cfg.TLS.CertFile = certFile
	}
	if keyFile := os.Getenv("TLS_KEY_FILE"); keyFile != "" {
		cfg.TLS.KeyFile = keyFile
	}
	if caFile := os.Getenv("TLS_CA_FILE"); caFile != "" {
		cfg.TLS.CAFile = caFile
	}

	// Manager configuration
	if managerURLs := os.Getenv("MANAGER_URLS"); managerURLs != "" {
		cfg.Manager.URLs = ParseManagerURLs(managerURLs)
	}
	if timeout := os.Getenv("MANAGER_TIMEOUT"); timeout != "" {
		fmt.Sscanf(timeout, "%d", &cfg.Manager.TimeoutSeconds)
	}

	// Set defaults
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Manager.TimeoutSeconds == 0 {
		cfg.Manager.TimeoutSeconds = 5
	}

	return &cfg, nil
}

// ParseManagerURLs parses comma-separated manager URLs from environment variable
func ParseManagerURLs(urlsStr string) []string {
	if urlsStr == "" {
		return nil
	}

	var urls []string
	for _, url := range strings.Split(urlsStr, ",") {
		url = strings.TrimSpace(url)
		if url != "" {
			urls = append(urls, url)
		}
	}
	return urls
}

// GetDSN returns PostgreSQL DSN string
func (c *Config) GetDSN() string {
	user := c.Database.User
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

// GetManagerURLs returns the list of manager service URLs
func (c *Config) GetManagerURLs() []string {
	return c.Manager.URLs
}

// GetManagerTimeout returns the manager timeout in seconds
func (c *Config) GetManagerTimeout() int {
	return c.Manager.TimeoutSeconds
}
