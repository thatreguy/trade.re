package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	Database    DatabaseConfig    `yaml:"database"`
	RIndex      RIndexConfig      `yaml:"rindex"`
	Auth        AuthConfig        `yaml:"auth"`
	Liquidation LiquidationConfig `yaml:"liquidation"`
	Game        GameConfig        `yaml:"game"`
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// DatabaseConfig holds PostgreSQL connection settings
type DatabaseConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	Name           string `yaml:"name"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	MaxConnections int    `yaml:"max_connections"`
}

// ConnectionString returns the PostgreSQL connection string
func (d DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, d.Name,
	)
}

// RIndexConfig holds R.index instrument settings
type RIndexConfig struct {
	StartingPrice decimal.Decimal `yaml:"starting_price"`
	TickSize      decimal.Decimal `yaml:"tick_size"`
	MinOrderSize  decimal.Decimal `yaml:"min_order_size"`
	MaxLeverage   int             `yaml:"max_leverage"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	JWTSecret        string `yaml:"jwt_secret"`
	TokenExpiryHours int    `yaml:"token_expiry_hours"`
	APIKeyLength     int    `yaml:"api_key_length"`
}

// LiquidationConfig holds liquidation engine settings
type LiquidationConfig struct {
	CheckIntervalMs      int                `yaml:"check_interval_ms"`
	InsuranceFundInitial decimal.Decimal    `yaml:"insurance_fund_initial"`
	MaintenanceMargins   MaintenanceMargins `yaml:"maintenance_margins"`
}

// MaintenanceMargins by leverage tier
type MaintenanceMargins struct {
	Conservative decimal.Decimal `yaml:"conservative"` // 1-10x
	Moderate     decimal.Decimal `yaml:"moderate"`     // 11-50x
	Aggressive   decimal.Decimal `yaml:"aggressive"`   // 51-100x
	Degen        decimal.Decimal `yaml:"degen"`        // 101-150x
}

// GetMarginForLeverage returns the maintenance margin for a given leverage
func (m MaintenanceMargins) GetMarginForLeverage(leverage int) decimal.Decimal {
	switch {
	case leverage <= 10:
		return m.Conservative
	case leverage <= 50:
		return m.Moderate
	case leverage <= 100:
		return m.Aggressive
	default:
		return m.Degen
	}
}

// GameConfig holds game-specific settings
type GameConfig struct {
	StartingBalance decimal.Decimal `yaml:"starting_balance"`
	CurrencySymbol  string          `yaml:"currency_symbol"`
}

// Load reads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	// Expand environment variables
	content := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	// Override with environment variables
	if pwd := os.Getenv("DB_PASSWORD"); pwd != "" {
		cfg.Database.Password = pwd
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.Auth.JWTSecret = secret
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks configuration for required fields
func (c *Config) Validate() error {
	var errs []string

	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		errs = append(errs, "server.port must be 1-65535")
	}

	if c.RIndex.MaxLeverage < 1 || c.RIndex.MaxLeverage > 150 {
		errs = append(errs, "rindex.max_leverage must be 1-150")
	}

	if c.RIndex.StartingPrice.LessThanOrEqual(decimal.Zero) {
		errs = append(errs, "rindex.starting_price must be positive")
	}

	if len(c.Auth.JWTSecret) > 0 && len(c.Auth.JWTSecret) < 32 {
		errs = append(errs, "auth.jwt_secret must be at least 32 characters")
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(errs, "; "))
	}

	return nil
}

// LoadOrDefault loads config from path or returns defaults
func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		// Return sensible defaults for development
		return &Config{
			Server: ServerConfig{
				Port: 8080,
				Host: "0.0.0.0",
			},
			Database: DatabaseConfig{
				Host:           "localhost",
				Port:           5432,
				Name:           "tradere",
				User:           "tradere",
				MaxConnections: 25,
			},
			RIndex: RIndexConfig{
				StartingPrice: decimal.NewFromInt(1000),
				TickSize:      decimal.NewFromFloat(0.01),
				MinOrderSize:  decimal.NewFromFloat(0.001),
				MaxLeverage:   150,
			},
			Auth: AuthConfig{
				TokenExpiryHours: 24,
				APIKeyLength:     32,
			},
			Liquidation: LiquidationConfig{
				CheckIntervalMs:      100,
				InsuranceFundInitial: decimal.NewFromInt(1000000),
				MaintenanceMargins: MaintenanceMargins{
					Conservative: decimal.NewFromFloat(0.005),
					Moderate:     decimal.NewFromFloat(0.01),
					Aggressive:   decimal.NewFromFloat(0.02),
					Degen:        decimal.NewFromFloat(0.05),
				},
			},
			Game: GameConfig{
				StartingBalance: decimal.NewFromInt(10000),
				CurrencySymbol:  "$",
			},
		}
	}
	return cfg
}
