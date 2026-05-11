package session

import (
	"encoding/base64"
	"fmt"

	"github.com/ricardoalcantara/min-idp/internal/config"
	localcrypto "github.com/ricardoalcantara/min-idp/internal/crypto"
)

type CookieTokenService struct {
	masterKey []byte
}

func NewCookieTokenService(cfg *config.Config) (*CookieTokenService, error) {
	key, err := localcrypto.DecodeMasterKey(cfg.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("cookie token: %w", err)
	}
	return &CookieTokenService{masterKey: key}, nil
}

// Encode encrypts a JWT string with AES-GCM and returns a base64url-encoded ciphertext.
func (s *CookieTokenService) Encode(jwt string) (string, error) {
	ciphertext, err := localcrypto.Encrypt(s.masterKey, []byte(jwt))
	if err != nil {
		return "", fmt.Errorf("cookie token encode: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

// Decode decrypts a base64url-encoded cookie value back to the original JWT string.
func (s *CookieTokenService) Decode(cookie string) (string, error) {
	ciphertext, err := base64.RawURLEncoding.DecodeString(cookie)
	if err != nil {
		return "", fmt.Errorf("cookie token decode: invalid base64: %w", err)
	}
	plaintext, err := localcrypto.Decrypt(s.masterKey, ciphertext)
	if err != nil {
		return "", fmt.Errorf("cookie token decode: decryption failed: %w", err)
	}
	return string(plaintext), nil
}
