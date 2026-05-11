package session

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	localcrypto "github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
)

type contextKey struct{}

// SessionClaims holds all identity information extracted from a validated JWT.
// Populated from JWT claims so controllers never need a DB hit for auth.
type SessionClaims struct {
	SessionUUID string
	UserUUID    string
	UserID      uint
	Email       string
	Roles       []string
	ExpiresAt   time.Time
}

// HasRole returns true if the given permission name is in the claims.
func (c *SessionClaims) HasRole(perm string) bool {
	for _, r := range c.Roles {
		if r == perm {
			return true
		}
	}
	return false
}

// APIMiddleware validates the session from either:
//  1. Session cookie (encrypted JWT) — browser clients
//  2. Authorization: Bearer <jwt> header — API clients
//
// After JWT signature verification it checks the KV revocation list.
// Populates SessionClaims in context; does not abort on missing auth.
func (s *SessionService) APIMiddleware(
	cookieName string,
	ks keystore.KeyStore,
	kv kvstore.KVStore,
	cookieToken *CookieTokenService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawJWT := extractJWT(c, cookieName, cookieToken)
		if rawJWT != "" {
			if claims, err := validateJWT(c.Request.Context(), rawJWT, ks, kv); err == nil {
				c.Set("session", claims)
				c.Request = c.Request.WithContext(
					context.WithValue(c.Request.Context(), contextKey{}, claims),
				)
			}
		}
		c.Next()
	}
}

// CookieMiddleware validates the session from the cookie only.
// Used by browser SSO flows (OIDC authorize, SAML SSO).
func (s *SessionService) CookieMiddleware(cookieName string, ks keystore.KeyStore, kv kvstore.KVStore, cookieToken *CookieTokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cookie, err := c.Cookie(cookieName); err == nil && cookie != "" {
			if rawJWT, err := cookieToken.Decode(cookie); err == nil {
				if claims, err := validateJWT(c.Request.Context(), rawJWT, ks, kv); err == nil {
					c.Set("session", claims)
					c.Request = c.Request.WithContext(
						context.WithValue(c.Request.Context(), contextKey{}, claims),
					)
				}
			}
		}
		c.Next()
	}
}

func extractJWT(c *gin.Context, cookieName string, cookieToken *CookieTokenService) string {
	// Cookie first (browser clients with encrypted JWT)
	if cookie, err := c.Cookie(cookieName); err == nil && cookie != "" {
		if rawJWT, err := cookieToken.Decode(cookie); err == nil {
			return rawJWT
		}
	}
	// Authorization: Bearer <jwt> (API clients)
	if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

func validateJWT(ctx context.Context, tokenStr string, ks keystore.KeyStore, kv kvstore.KVStore) (*SessionClaims, error) {
	keys, err := ks.PublicKeys(ctx, keystore_entities.ProtocolOIDC)
	if err != nil {
		return nil, err
	}

	// Build a kid→publicKey lookup map
	keyMap := make(map[string]interface{}, len(keys))
	for _, k := range keys {
		pub, err := localcrypto.ParsePublicKeyPEM([]byte(k.PublicKey))
		if err != nil {
			continue
		}
		keyMap[k.KID] = pub
	}

	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("session: missing kid in JWT header")
		}
		pub, ok := keyMap[kid]
		if !ok {
			return nil, errors.New("session: unknown kid")
		}
		switch t.Method.(type) {
		case *jwt.SigningMethodECDSA:
			if _, ok := pub.(*ecdsa.PublicKey); !ok {
				return nil, errors.New("session: key type mismatch")
			}
		case *jwt.SigningMethodRSA:
			if _, ok := pub.(*rsa.PublicKey); !ok {
				return nil, errors.New("session: key type mismatch")
			}
		default:
			return nil, errors.New("session: unsupported signing method")
		}
		return pub, nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("session: invalid JWT")
	}

	mapClaims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("session: invalid claims")
	}

	sid, _ := mapClaims["sid"].(string)
	if sid == "" {
		return nil, errors.New("session: missing sid claim")
	}

	// KV revocation check
	if _, err := kv.Get(ctx, "revoked:"+sid); err == nil {
		return nil, errors.New("session: revoked")
	}

	userUUID, _ := mapClaims["sub"].(string)
	email, _ := mapClaims["email"].(string)

	var userID uint
	if uid, ok := mapClaims["uid"].(float64); ok {
		userID = uint(uid)
	}

	var roles []string
	if raw, ok := mapClaims["roles"].([]interface{}); ok {
		for _, r := range raw {
			if s, ok := r.(string); ok {
				roles = append(roles, s)
			}
		}
	}

	var expiresAt time.Time
	if exp, ok := mapClaims["exp"].(float64); ok {
		expiresAt = time.Unix(int64(exp), 0)
	}

	return &SessionClaims{
		SessionUUID: sid,
		UserUUID:    userUUID,
		UserID:      userID,
		Email:       email,
		Roles:       roles,
		ExpiresAt:   expiresAt,
	}, nil
}

func RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("session"); !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, web.NewErrorDto(errors.New("authentication required")))
			return
		}
		c.Next()
	}
}

func FromContext(c *gin.Context) *SessionClaims {
	raw, _ := c.Get("session")
	claims, _ := raw.(*SessionClaims)
	return claims
}
