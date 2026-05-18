package config

import "github.com/caarlos0/env/v10"

type Config struct {
	Port           string `env:"PORT" envDefault:"8081"`
	Env            string `env:"ENV" envDefault:"development"`
	DatabaseURL    string `env:"DATABASE_URL" envDefault:"postgres://gbspos:gbspos_secret@localhost:5432/gbs_pos?sslmode=disable"`
	JWTSecret      string `env:"JWT_SECRET" envDefault:"your-super-secret-jwt-key-minimum-32-characters"`
	JWTExpiryHours int    `env:"JWT_EXPIRY_HOURS" envDefault:"24"`
	LogLevel       string `env:"LOG_LEVEL" envDefault:"debug"`
	UploadDir      string `env:"UPLOAD_DIR" envDefault:"./uploads/ads"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
