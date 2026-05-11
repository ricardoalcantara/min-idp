package authn

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MintSessionJWT signs a session JWT with the given key and returns the token string.
// The JWT is suitable for use as an API access token or encrypted browser cookie.
func MintSessionJWT(
	key crypto.PrivateKey,
	kid, alg, issuer string,
	userID uint,
	userUUID, sessionUUID, email string,
	roles []string,
	expiry time.Duration,
) (string, error) {
	now := time.Now().UTC()

	claims := jwt.MapClaims{
		"iss":   issuer,
		"sub":   userUUID,
		"uid":   userID,
		"sid":   sessionUUID,
		"email": email,
		"roles": roles,
		"iat":   now.Unix(),
		"exp":   now.Add(expiry).Unix(),
	}

	var signingMethod jwt.SigningMethod
	switch alg {
	case "ES256":
		signingMethod = jwt.SigningMethodES256
	case "RS256":
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", fmt.Errorf("authn: unsupported signing algorithm %q", alg)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	token.Header["kid"] = kid

	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		return token.SignedString(k)
	case *rsa.PrivateKey:
		return token.SignedString(k)
	default:
		return "", fmt.Errorf("authn: unsupported key type %T", key)
	}
}
