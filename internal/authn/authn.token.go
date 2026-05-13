package authn

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SigningConfig groups the per-key metadata needed to mint a session token.
type SigningConfig struct {
	KID    string
	Issuer string
}

type sessionClaims struct {
	jwt.RegisteredClaims
	UserID      uint     `json:"uid"`
	SessionUUID string   `json:"sid"`
	Email       string   `json:"email"`
	Name        string   `json:"name,omitempty"`
	Roles       []string `json:"roles"`
}

// MintSessionJWT signs a session JWT for the given user. The signing algorithm
// is derived from the key type, eliminating alg/key mismatch at call time.
func MintSessionJWT(
	key crypto.PrivateKey,
	cfg SigningConfig,
	userID uint,
	userUUID, sessionUUID, email, name string,
	roles []string,
	expiry time.Duration,
) (string, error) {
	now := time.Now()

	claims := sessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.Issuer,
			Subject:   userUUID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
		UserID:      userID,
		SessionUUID: sessionUUID,
		Email:       email,
		Name:        name,
		Roles:       roles,
	}

	var signingMethod jwt.SigningMethod
	switch key.(type) {
	case *ecdsa.PrivateKey:
		signingMethod = jwt.SigningMethodES256
	case *rsa.PrivateKey:
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", fmt.Errorf("authn: unsupported key type %T", key)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	token.Header["kid"] = cfg.KID

	return token.SignedString(key)
}
