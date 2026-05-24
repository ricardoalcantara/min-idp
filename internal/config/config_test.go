package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateSMTP_Disabled(t *testing.T) {
	cfg := &Config{SMTPEnabled: false}
	assert.NoError(t, cfg.ValidateSMTP())
}

func TestConfig_ValidateSMTP_EnabledMissingHost(t *testing.T) {
	cfg := &Config{SMTPEnabled: true, SMTPFrom: "noreply@example.com"}
	err := cfg.ValidateSMTP()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MIN_IDP_SMTP_HOST")
}

func TestConfig_ValidateSMTP_EnabledMissingFrom(t *testing.T) {
	cfg := &Config{SMTPEnabled: true, SMTPHost: "smtp.example.com"}
	err := cfg.ValidateSMTP()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MIN_IDP_SMTP_FROM")
}

func TestConfig_ValidateSMTP_InvalidTLS(t *testing.T) {
	cfg := &Config{
		SMTPEnabled: true,
		SMTPHost:    "smtp.example.com",
		SMTPFrom:    "noreply@example.com",
		SMTPTLS:     "invalid",
	}
	err := cfg.ValidateSMTP()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MIN_IDP_SMTP_TLS")
}

func TestConfig_ValidateSMTP_Valid(t *testing.T) {
	cfg := &Config{
		SMTPEnabled: true,
		SMTPHost:    "smtp.example.com",
		SMTPFrom:    "noreply@example.com",
		SMTPTLS:     "starttls",
	}
	assert.NoError(t, cfg.ValidateSMTP())
}
