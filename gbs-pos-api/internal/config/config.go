package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Port            string `env:"PORT"              envDefault:"8080"`
	Env             string `env:"ENV"               envDefault:"development"`
	DatabaseURL     string `env:"DATABASE_URL"      envDefault:"postgres://gbspos:gbspos_secret@localhost:5432/gbs_pos?sslmode=disable"`
	MigrationsPath  string `env:"MIGRATIONS_PATH"   envDefault:""`
	JWTSecret       string `env:"JWT_SECRET"`
	JWTExpiryHours  int    `env:"JWT_EXPIRY_HOURS"  envDefault:"24"`
	LogLevel        string `env:"LOG_LEVEL"         envDefault:"debug"`
	UploadDir       string `env:"UPLOAD_DIR"        envDefault:"./uploads/ads"`
	KeycloakBaseURL string `env:"KEYCLOAK_BASE_URL" envDefault:""`
	KeycloakRealm   string `env:"KEYCLOAK_REALM"    envDefault:""`
	EnableDemoAuth  bool   `env:"ENABLE_DEMO_AUTH"  envDefault:"false"`
}

func (c *Config) UseKeycloak() bool {
	return c.KeycloakBaseURL != "" && c.KeycloakRealm != ""
}

func (c *Config) KeycloakJWKSURL() string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", c.KeycloakBaseURL, c.KeycloakRealm)
}

func (c *Config) Validate() error {
	if c.UseKeycloak() {
		if c.EnableDemoAuth && (c.JWTSecret == "" || len(c.JWTSecret) < 32) {
			return fmt.Errorf("JWT_SECRET is required and must be at least 32 characters when ENABLE_DEMO_AUTH is true")
		}
		return nil
	}

	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required and must be at least 32 characters")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return nil
}

func Load() (*Config, error) {
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load(".env")

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
