package oidc

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	localcrypto "github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	oidc_dto "github.com/ricardoalcantara/min-idp/internal/protocol/oidc/dto"
	oidc_entities "github.com/ricardoalcantara/min-idp/internal/protocol/oidc/entities"
	"github.com/ricardoalcantara/min-idp/internal/rbac"
	"github.com/ricardoalcantara/min-idp/internal/sp"
	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
)

var (
	ErrInvalidClient       = errors.New("invalid client")
	ErrInvalidGrant        = errors.New("invalid grant")
	ErrUnsupportedGrant    = errors.New("unsupported grant type")
	ErrInvalidRedirectURI  = errors.New("invalid redirect uri")
	ErrInvalidRequest      = errors.New("invalid request")
	ErrAccessDenied        = errors.New("access denied")
)

type OAuthTokenRepository interface {
	CreateToken(token *oidc_entities.OAuthToken) error
	FindByHash(hash string) (*oidc_entities.OAuthToken, error)
	RevokeToken(hash string) error
}

type AuthCodeData struct {
	ClientID            string `json:"client_id"`
	UserID              uint   `json:"user_id"`
	UserUUID            string `json:"user_uuid"`
	SessionUUID         string `json:"session_uuid"`
	RedirectURI         string `json:"redirect_uri"`
	Scope               string `json:"scope"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	Nonce               string `json:"nonce"`
}

type OIDCService struct {
	tokenRepo OAuthTokenRepository
	spRepo    sp.SPRepository
	kv        kvstore.KVStore
	ks        keystore.KeyStore
	cfg       *config.Config
	rbacSvc   *rbac.RBACService
}

func NewOIDCService(
	tokenRepo OAuthTokenRepository,
	spRepo sp.SPRepository,
	kv kvstore.KVStore,
	ks keystore.KeyStore,
	cfg *config.Config,
	rbacSvc *rbac.RBACService,
) *OIDCService {
	return &OIDCService{
		tokenRepo: tokenRepo,
		spRepo:    spRepo,
		kv:        kv,
		ks:        ks,
		cfg:       cfg,
		rbacSvc:   rbacSvc,
	}
}

// GenerateAuthCode generates an authorization code, stores it in KV, and returns the code string.
func (s *OIDCService) GenerateAuthCode(ctx context.Context, data AuthCodeData) (string, error) {
	code := uuid.NewString()
	hash := sha256.Sum256([]byte(code))
	codeHash := base64.RawURLEncoding.EncodeToString(hash[:])

	val, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// Store code in KV with a 5-minute TTL
	err = s.kv.Set(ctx, "oauth_code:"+codeHash, val, 5*time.Minute)
	if err != nil {
		return "", err
	}

	return code, nil
}

// ValidateClient returns the OIDC client if credentials (if any) and redirect URI match.
func (s *OIDCService) ValidateClient(clientID, clientSecret, redirectURI string) (*sp_entities.OIDCClient, error) {
	client, err := s.spRepo.FindOIDCClientByClientID(clientID)
	if err != nil {
		return nil, ErrInvalidClient
	}

	if clientSecret != "" {
		if err := localcrypto.VerifyPassword(client.ClientSecretHash, clientSecret); err != nil {
			return nil, ErrInvalidClient
		}
	}

	if redirectURI != "" {
		validURIs := sp_repositories.UnmarshalStringSlice(client.RedirectURIs)
		found := false
		for _, uri := range validURIs {
			if uri == redirectURI {
				found = true
				break
			}
		}
		if !found {
			return nil, ErrInvalidRedirectURI
		}
	}

	return client, nil
}

// ExchangeCode performs the auth code grant exchange.
func (s *OIDCService) ExchangeCode(ctx context.Context, req oidc_dto.TokenRequest, client *sp_entities.OIDCClient) (*oidc_dto.TokenResponse, error) {
	hash := sha256.Sum256([]byte(req.Code))
	codeHash := base64.RawURLEncoding.EncodeToString(hash[:])

	val, err := s.kv.Get(ctx, "oauth_code:"+codeHash)
	if err != nil {
		return nil, ErrInvalidGrant
	}
	// Delete immediately after fetching to prevent reuse
	_ = s.kv.Delete(ctx, "oauth_code:"+codeHash)

	var data AuthCodeData
	if err := json.Unmarshal(val, &data); err != nil {
		return nil, ErrInvalidGrant
	}

	if data.ClientID != client.ClientID {
		return nil, ErrInvalidGrant
	}
	if data.RedirectURI != req.RedirectURI {
		return nil, ErrInvalidGrant
	}

	if client.PKCERequired || data.CodeChallenge != "" {
		if req.CodeVerifier == "" {
			return nil, ErrInvalidRequest
		}
		if data.CodeChallengeMethod == "S256" {
			h := sha256.Sum256([]byte(req.CodeVerifier))
			enc := base64.RawURLEncoding.EncodeToString(h[:])
			if enc != data.CodeChallenge {
				return nil, ErrInvalidGrant
			}
		} else if data.CodeChallengeMethod == "plain" {
			if req.CodeVerifier != data.CodeChallenge {
				return nil, ErrInvalidGrant
			}
		}
	}

	return s.issueTokens(ctx, client, data.UserID, data.UserUUID, data.SessionUUID, data.Scope, data.Nonce, nil)
}

// ExchangeRefreshToken performs the refresh token grant.
func (s *OIDCService) ExchangeRefreshToken(ctx context.Context, req oidc_dto.TokenRequest, client *sp_entities.OIDCClient) (*oidc_dto.TokenResponse, error) {
	hash := sha256.Sum256([]byte(req.RefreshToken))
	tokenHash := base64.RawURLEncoding.EncodeToString(hash[:])

	token, err := s.tokenRepo.FindByHash(tokenHash)
	if err != nil || token.Type != "refresh" || token.RevokedAt != nil || token.ExpiresAt.Before(time.Now()) {
		return nil, ErrInvalidGrant
	}

	if token.ClientID != client.ClientID {
		return nil, ErrInvalidGrant
	}

	// Revoke old refresh token (refresh token rotation)
	_ = s.tokenRepo.RevokeToken(tokenHash)

	// Fetch user UUID (would need UserRepo in real life or we just store userUUID in token)
	// For now we'll just pass empty or we can add UserUUID to OAuthToken. Let's assume we need to pass userUUID.
	// We can decode the user UUID from session or fetch it. We will just pass an empty string for now, but
	// ID token might need it.

	return s.issueTokens(ctx, client, token.UserID, "", token.SessionUUID, token.Scope, "", &token.ID)
}

func (s *OIDCService) issueTokens(
	ctx context.Context,
	client *sp_entities.OIDCClient,
	userID uint,
	userUUID string,
	sessionUUID string,
	scope string,
	nonce string,
	parentID *uint,
) (*oidc_dto.TokenResponse, error) {
	now := time.Now()
	// Scopes can dictate what is issued
	roles, _ := s.rbacSvc.GetUserPermissions(userID)

	// Issue Access Token as JWT
	accessJTI := uuid.NewString()
	accessToken, err := s.mintAccessToken(ctx, client.ClientID, accessJTI, userID, userUUID, roles, scope, 1*time.Hour)
	if err != nil {
		return nil, err
	}

	// Save Access Token JTI
	err = s.tokenRepo.CreateToken(&oidc_entities.OAuthToken{
		Type:        "access",
		TokenHash:   accessJTI, // storing JTI directly
		ClientID:    client.ClientID,
		UserID:      userID,
		SessionUUID: sessionUUID,
		Scope:       scope,
		ExpiresAt:   now.Add(1 * time.Hour),
		ParentID:    parentID,
	})
	if err != nil {
		return nil, err
	}

	// Issue Refresh Token as opaque
	rawRefresh := uuid.NewString()
	refreshHashRaw := sha256.Sum256([]byte(rawRefresh))
	refreshHash := base64.RawURLEncoding.EncodeToString(refreshHashRaw[:])

	err = s.tokenRepo.CreateToken(&oidc_entities.OAuthToken{
		Type:        "refresh",
		TokenHash:   refreshHash,
		ClientID:    client.ClientID,
		UserID:      userID,
		SessionUUID: sessionUUID,
		Scope:       scope,
		ExpiresAt:   now.Add(24 * 7 * time.Hour),
		ParentID:    parentID,
	})
	if err != nil {
		return nil, err
	}

	// Issue ID Token if openid scope is present
	var idTokenStr string
	if strings.Contains(scope, "openid") {
		idTokenStr, err = s.mintIDToken(ctx, client.ClientID, userUUID, nonce, roles, 1*time.Hour)
		if err != nil {
			return nil, err
		}
	}

	return &oidc_dto.TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: rawRefresh,
		IDToken:      idTokenStr,
	}, nil
}

// Introspect checks if an opaque token or access JWT is valid.
func (s *OIDCService) Introspect(ctx context.Context, tokenStr string) (*oidc_dto.IntrospectResponse, error) {
	// Try to decode as JWT
	claims, err := s.decodeJWT(ctx, tokenStr)
	if err != nil {
		return &oidc_dto.IntrospectResponse{Active: false}, nil
	}

	jti, _ := claims["jti"].(string)
	if jti == "" {
		return &oidc_dto.IntrospectResponse{Active: false}, nil
	}

	token, err := s.tokenRepo.FindByHash(jti)
	if err != nil || token.RevokedAt != nil || token.ExpiresAt.Before(time.Now()) {
		return &oidc_dto.IntrospectResponse{Active: false}, nil
	}

	exp, _ := claims["exp"].(float64)
	iat, _ := claims["iat"].(float64)
	sub, _ := claims["sub"].(string)
	clientID, _ := claims["client_id"].(string)

	return &oidc_dto.IntrospectResponse{
		Active:    true,
		Scope:     token.Scope,
		ClientID:  clientID,
		TokenType: "Bearer",
		Exp:       int64(exp),
		Iat:       int64(iat),
		Sub:       sub,
	}, nil
}

func (s *OIDCService) Revoke(ctx context.Context, tokenStr string) error {
	// If it's a JWT, revoke by JTI
	if claims, err := s.decodeJWT(ctx, tokenStr); err == nil {
		if jti, ok := claims["jti"].(string); ok && jti != "" {
			return s.tokenRepo.RevokeToken(jti)
		}
	}
	// Otherwise treat as opaque refresh token
	hash := sha256.Sum256([]byte(tokenStr))
	tokenHash := base64.RawURLEncoding.EncodeToString(hash[:])
	return s.tokenRepo.RevokeToken(tokenHash)
}

func (s *OIDCService) GetUserInfo(ctx context.Context, accessToken string) (*oidc_dto.UserInfoResponse, error) {
	claims, err := s.decodeJWT(ctx, accessToken)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	jti, _ := claims["jti"].(string)
	token, err := s.tokenRepo.FindByHash(jti)
	if err != nil || token.RevokedAt != nil || token.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("token revoked or expired")
	}

	sub, _ := claims["sub"].(string)
	var roles string
	if rArr, ok := claims["roles"].([]interface{}); ok {
		var strRoles []string
		for _, v := range rArr {
			if s, ok := v.(string); ok {
				strRoles = append(strRoles, s)
			}
		}
		roles = strings.Join(strRoles, ",")
	}

	return &oidc_dto.UserInfoResponse{
		Sub:   sub,
		Roles: roles,
	}, nil
}

// Mint access token as JWT
func (s *OIDCService) mintAccessToken(ctx context.Context, clientID, jti string, userID uint, userUUID string, roles []string, scope string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":       s.cfg.ExternalURL,
		"sub":       userUUID,
		"client_id": clientID,
		"jti":       jti,
		"scope":     scope,
		"roles":     roles,
		"iat":       jwt.NewNumericDate(now),
		"exp":       jwt.NewNumericDate(now.Add(expiry)),
	}

	return s.signClaims(ctx, claims)
}

func (s *OIDCService) mintIDToken(ctx context.Context, clientID, userUUID, nonce string, roles []string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   s.cfg.ExternalURL,
		"sub":   userUUID,
		"aud":   clientID,
		"roles": roles,
		"iat":   jwt.NewNumericDate(now),
		"exp":   jwt.NewNumericDate(now.Add(expiry)),
	}

	if nonce != "" {
		claims["nonce"] = nonce
	}

	return s.signClaims(ctx, claims)
}

func (s *OIDCService) signClaims(ctx context.Context, claims jwt.Claims) (string, error) {
	key, meta, err := s.ks.ActivePrivateKey(ctx, keystore_entities.ProtocolOIDC)
	if err != nil {
		return "", err
	}

	var signingMethod jwt.SigningMethod
	switch key.(type) {
	case *ecdsa.PrivateKey:
		signingMethod = jwt.SigningMethodES256
	case *rsa.PrivateKey:
		signingMethod = jwt.SigningMethodRS256
	default:
		return "", fmt.Errorf("oidc: unsupported key type %T", key)
	}

	token := jwt.NewWithClaims(signingMethod, claims)
	token.Header["kid"] = meta.KID

	return token.SignedString(key)
}

func (s *OIDCService) decodeJWT(ctx context.Context, tokenStr string) (jwt.MapClaims, error) {
	keys, err := s.ks.PublicKeys(ctx, keystore_entities.ProtocolOIDC)
	if err != nil {
		return nil, err
	}

	keyMap := make(map[string]interface{}, len(keys))
	for _, k := range keys {
		pub, err := localcrypto.ParsePublicKeyPEM([]byte(k.PublicKey))
		if err != nil {
			continue
		}
		keyMap[k.KID] = pub
	}

	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		pub, ok := keyMap[kid]
		if !ok {
			return nil, errors.New("unknown kid")
		}
		return pub, nil
	})

	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
