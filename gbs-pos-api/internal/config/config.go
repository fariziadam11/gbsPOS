package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type Config struct {
	Port             string `env:"PORT"              envDefault:"8080"`
	Env              string `env:"ENV"               envDefault:"development"`
	DatabaseURL      string `env:"DATABASE_URL"      envDefault:"postgres://gbspos:gbspos_secret@localhost:5432/gbs_pos?sslmode=disable"`
	MigrationsPath   string `env:"MIGRATIONS_PATH"   envDefault:""`
	JWTSecret        string `env:"JWT_SECRET"`
	JWTExpiryHours   int    `env:"JWT_EXPIRY_HOURS"  envDefault:"24"`
	LogLevel         string `env:"LOG_LEVEL"         envDefault:"debug"`
	UploadDir        string `env:"UPLOAD_DIR"        envDefault:"./uploads/ads"`
	KeycloakBaseURL  string `env:"KEYCLOAK_BASE_URL" envDefault:""`
	KeycloakRealm    string `env:"KEYCLOAK_REALM"    envDefault:""`
	EnableDemoAuth   bool   `env:"ENABLE_DEMO_AUTH"  envDefault:"false"`
	WSAllowedOrigins string `env:"WS_ALLOWED_ORIGINS" envDefault:""`
	// PublicBaseURL is the externally reachable base URL (no trailing slash).
	// Used to build public URLs embedded in QR codes (e.g. the fuel dispenser
	// authorization scan URL). Defaults to the production API domain.
	PublicBaseURL string `env:"PUBLIC_BASE_URL" envDefault:"https://api-pos.armmada.id"`

	// SumoPod QRIS Payment Gateway
	SumoPodAPIURL       string `env:"SUMOPOD_API_URL"        envDefault:"https://api-pay-sandbox.sumopod.com/api/v1"`
	SumoPodAPIKey       string `env:"SUMOPOD_API_KEY"        envDefault:""`
	SumoPodWebhookToken string `env:"SUMOPOD_WEBHOOK_TOKEN"  envDefault:""`
	SumoPodSuccessURL   string `env:"SUMOPOD_SUCCESS_URL"    envDefault:""`
	SumoPodCancelURL    string `env:"SUMOPOD_CANCEL_URL"     envDefault:""`
	SumoPodExpiresHours int    `env:"SUMOPOD_EXPIRES_HOURS"  envDefault:"24"`

	// QRIS Direct (Static to Dynamic)
	// Static QRIS string for Gojek/Mama Tari
	QrisDirectStaticQRIS              string `env:"QRIS_DIRECT_STATIC_QRIS" envDefault:"00020101021126610014COM.GO-JEK.WWW01189360091434374848210210G4374848210303UMI51440014ID.CO.QRIS.WWW0215ID10265153412990303UMI5204581253033605802ID5925Snack Kering Mama Tari, K6013JAKARTA PUSAT61051064062140703A01110362163042807"`
	QrisDirectMerchantName            string `env:"QRIS_DIRECT_MERCHANT_NAME" envDefault:"Snack Kering Mama Tari, K"`
	QrisDirectMerchantCity            string `env:"QRIS_DIRECT_MERCHANT_CITY" envDefault:"JAKARTA PUSAT"`
	QrisDirectProvider                string `env:"QRIS_DIRECT_PROVIDER" envDefault:"GoPay"`
	QrisDirectExpiresMinutes          int    `env:"QRIS_DIRECT_EXPIRES_MINUTES" envDefault:"15"`
	QrisDirectSkipCRCValidate         bool   `env:"QRIS_DIRECT_SKIP_CRC_VALIDATE" envDefault:"true"`
	QrisDirectAutoConfirmDelaySeconds int    `env:"QRIS_DIRECT_AUTO_CONFIRM_DELAY_SECONDS" envDefault:"3"`

	BtnSnapEnabled        bool    `env:"BTN_SNAP_ENABLED" envDefault:"false"`
	BtnSnapBaseURL        string  `env:"BTN_SNAP_BASE_URL" envDefault:"https://devapi.btn.co.id"`
	BtnSnapClientKey      string  `env:"BTN_SNAP_CLIENT_KEY"`
	BtnSnapPartnerID      string  `env:"BTN_SNAP_PARTNER_ID"`
	BtnSnapClientSecret   string  `env:"BTN_SNAP_CLIENT_SECRET"`
	BtnSnapPrivateKeyPath string  `env:"BTN_SNAP_PRIVATE_KEY_PATH"`
	BtnSnapChannelID      string  `env:"BTN_SNAP_CHANNEL_ID"`
	BtnSnapOrigin         string  `env:"BTN_SNAP_ORIGIN"`
	BtnSnapMerchantID     string  `env:"BTN_SNAP_MERCHANT_ID"`
	BtnSnapTerminalID     string  `env:"BTN_SNAP_TERMINAL_ID"`
	PosTaxRate            float64 `env:"POS_TAX_RATE" envDefault:"0.10"`
}

func (c *Config) UseKeycloak() bool {
	return c.KeycloakBaseURL != "" && c.KeycloakRealm != ""
}

func (c *Config) KeycloakJWKSURL() string {
	return fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", c.KeycloakBaseURL, c.KeycloakRealm)
}

func (c *Config) Validate() error {
	if c.PosTaxRate < 0 || c.PosTaxRate > 1 {
		return fmt.Errorf("POS_TAX_RATE must be between 0 and 1")
	}
	if c.BtnSnapEnabled {
		if c.BtnSnapClientKey == "" || c.BtnSnapPartnerID == "" || c.BtnSnapClientSecret == "" ||
			c.BtnSnapPrivateKeyPath == "" || c.BtnSnapChannelID == "" || c.BtnSnapOrigin == "" ||
			c.BtnSnapMerchantID == "" || c.BtnSnapTerminalID == "" {
			return fmt.Errorf("all BTN_SNAP credentials, merchant, terminal, channel, and origin values are required when BTN_SNAP_ENABLED is true")
		}
		baseURL, err := url.ParseRequestURI(c.BtnSnapBaseURL)
		if err != nil || baseURL.Scheme != "https" || baseURL.Host == "" || (baseURL.Path != "" && baseURL.Path != "/") || baseURL.RawQuery != "" || baseURL.Fragment != "" {
			return fmt.Errorf("BTN_SNAP_BASE_URL must be an HTTPS base URL without query or fragment")
		}
		if len(c.BtnSnapChannelID) != 5 {
			return fmt.Errorf("BTN_SNAP_CHANNEL_ID must contain exactly 5 digits")
		}
		for _, r := range c.BtnSnapChannelID {
			if r < '0' || r > '9' {
				return fmt.Errorf("BTN_SNAP_CHANNEL_ID must contain exactly 5 digits")
			}
		}
	}

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

	// Trim whitespace from QRIS string (handles accidental newlines/spaces in .env)
	cfg.QrisDirectStaticQRIS = strings.TrimSpace(cfg.QrisDirectStaticQRIS)

	// Trim trailing slash from the public base URL (kept in config.Load, not
	// Validate, so consumers can rely on a clean join with path segments).
	cfg.PublicBaseURL = strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/")

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
