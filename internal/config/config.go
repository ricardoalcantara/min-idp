package config

import (
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

	// AdminPassword pins the initial admin password on first bootstrap.
	// Leave empty to generate a random password (recommended for production).
	AdminPassword string `env:"MIN_IDP_ADMIN_PASSWORD"`

	FeatureAPIRegistration bool `env:"MIN_IDP_FEATURE_API_REGISTRATION" envDefault:"false"`
	FeatureAPILogin        bool `env:"MIN_IDP_FEATURE_API_LOGIN"        envDefault:"true"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
