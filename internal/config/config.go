package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ExternalURL string `env:"MIN_IDP_EXTERNAL_URL" envRequired:"true"`
	DBDriver    string `env:"MIN_IDP_DB_DRIVER"    envDefault:"sqlite"`
	DBDSN       string `env:"MIN_IDP_DB_DSN"       envRequired:"true"`
	MasterKey   string `env:"MIN_IDP_MASTER_KEY"   envRequired:"true"`

	KVDriver   string `env:"MIN_IDP_KV_DRIVER" envDefault:"db"`
	KVRedisURL string `env:"MIN_IDP_KV_REDIS_URL"`

	SessionCookie string        `env:"MIN_IDP_SESSION_COOKIE" envDefault:"min_idp_session"`
	SessionTTL    time.Duration `env:"MIN_IDP_SESSION_TTL"    envDefault:"12h"`
	SessionIdle   time.Duration `env:"MIN_IDP_SESSION_IDLE"   envDefault:"1h"`

	PasswordResetTTL time.Duration `env:"MIN_IDP_PASSWORD_RESET_TTL" envDefault:"15m"`

	// AdminPassword pins the initial admin password on first bootstrap.
	// Leave empty to generate a random password (recommended for production).
	AdminPassword string `env:"MIN_IDP_ADMIN_PASSWORD"`

	FeatureAPIRegistration bool `env:"MIN_IDP_FEATURE_API_REGISTRATION" envDefault:"false"`
	FeatureAPILogin        bool `env:"MIN_IDP_FEATURE_API_LOGIN"        envDefault:"true"`

	SMTPEnabled  bool   `env:"MIN_IDP_SMTP_ENABLED"   envDefault:"false"`
	SMTPHost     string `env:"MIN_IDP_SMTP_HOST"`
	SMTPPort     int    `env:"MIN_IDP_SMTP_PORT"      envDefault:"587"`
	SMTPUsername string `env:"MIN_IDP_SMTP_USERNAME"`
	SMTPPassword string `env:"MIN_IDP_SMTP_PASSWORD"`
	SMTPFrom     string `env:"MIN_IDP_SMTP_FROM"`
	SMTPFromName string `env:"MIN_IDP_SMTP_FROM_NAME" envDefault:"min-idp"`
	SMTPTLS      string `env:"MIN_IDP_SMTP_TLS"       envDefault:"starttls"`
}

func (c *Config) ValidateSMTP() error {
	if !c.SMTPEnabled {
		return nil
	}
	if c.SMTPHost == "" {
		return fmt.Errorf("MIN_IDP_SMTP_HOST is required when MIN_IDP_SMTP_ENABLED=true")
	}
	if c.SMTPFrom == "" {
		return fmt.Errorf("MIN_IDP_SMTP_FROM is required when MIN_IDP_SMTP_ENABLED=true")
	}
	switch c.SMTPTLS {
	case "starttls", "ssl", "none":
	default:
		return fmt.Errorf("MIN_IDP_SMTP_TLS must be starttls, ssl, or none")
	}
	return nil
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	if err := cfg.ValidateSMTP(); err != nil {
		return nil, err
	}
	return cfg, nil
}
