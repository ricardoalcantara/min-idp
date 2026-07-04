package oidc

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/ricardoalcantara/min-idp/internal/jwtutil"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/config"
	localcrypto "github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	oidc_dto "github.com/ricardoalcantara/min-idp/internal/protocol/oidc/dto"
	"github.com/ricardoalcantara/min-idp/internal/session"
	"github.com/ricardoalcantara/min-idp/internal/sp"
	"github.com/ricardoalcantara/min-idp/internal/views"
)

type OIDCController struct {
	ks      keystore.KeyStore
	kv      kvstore.KVStore
	service *OIDCService
	gate    *sp.SPGateService
	cfg     *config.Config
	log     *slog.Logger
	issuer  string
}

func NewOIDCController(ks keystore.KeyStore, kv kvstore.KVStore, service *OIDCService, gate *sp.SPGateService, cfg *config.Config, log *slog.Logger) *OIDCController {
	return &OIDCController{ks: ks, kv: kv, service: service, gate: gate, cfg: cfg, log: log, issuer: cfg.ExternalURL}
}

func (c *OIDCController) discovery(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, oidc_dto.Build(c.issuer))
}

func (c *OIDCController) jwks(ctx *gin.Context) {
	keys, err := c.ks.PublicKeys(ctx.Request.Context(), keystore_entities.ProtocolOIDC)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	jwks := localcrypto.JWKS{Keys: make([]localcrypto.JWK, 0, len(keys))}
	for _, k := range keys {
		pub, err := localcrypto.ParsePublicKeyPEM([]byte(k.PublicKey))
		if err != nil {
			continue
		}
		switch pk := pub.(type) {
		case *rsa.PublicKey:
			jwks.Keys = append(jwks.Keys, localcrypto.RSAPublicKeyToJWK(pk, k.KID, k.Algorithm))
		case *ecdsa.PublicKey:
			jwks.Keys = append(jwks.Keys, localcrypto.ECPublicKeyToJWK(pk, k.KID, k.Algorithm))
		}
	}
	ctx.JSON(http.StatusOK, jwks)
}

func (c *OIDCController) authorize(ctx *gin.Context) {
	clientID := ctx.Query("client_id")
	redirectURI := ctx.Query("redirect_uri")
	responseType := ctx.Query("response_type")
	scope := ctx.Query("scope")
	state := ctx.Query("state")
	codeChallenge := ctx.Query("code_challenge")
	codeChallengeMethod := ctx.Query("code_challenge_method")
	nonce := ctx.Query("nonce")

	if responseType != "code" {
		c.redirectError(ctx, redirectURI, "unsupported_response_type", state)
		return
	}

	client, err := c.service.ValidateClient(clientID, "", redirectURI)
	if err != nil {
		c.redirectError(ctx, redirectURI, "invalid_request", state)
		return
	}

	spEntity, err := c.service.spRepo.FindByID(client.SPID)
	if err != nil {
		c.redirectError(ctx, redirectURI, "server_error", state)
		return
	}

	// Check if user is logged in
	sess := session.FromContext(ctx)
	if sess == nil {
		// Not logged in. Stash request and go to login
		c.stashAndRedirectToLogin(ctx, ctx.Request.URL.String())
		return
	}

	// Check SSO permissions
	if err := c.gate.CanSSO(sess.UserID, spEntity); err != nil {
		c.log.Debug("SSO gate denied access", "user_id", sess.UserID, "sp_id", spEntity.ID, "error", err)
		c.redirectError(ctx, redirectURI, "access_denied", state)
		return
	}

	// Valid session and allowed. Generate code.
	code, err := c.service.GenerateAuthCode(ctx.Request.Context(), AuthCodeData{
		ClientID:            clientID,
		UserID:              sess.UserID,
		UserUUID:            sess.UserUUID,
		Email:               sess.Email,
		Username:            sess.Username,
		Name:                sess.Name,
		SessionUUID:         sess.SessionUUID,
		RedirectURI:         redirectURI,
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Nonce:               nonce,
	})
	if err != nil {
		c.redirectError(ctx, redirectURI, "server_error", state)
		return
	}

	c.redirectSuccess(ctx, redirectURI, code, state)
}

func (c *OIDCController) stashAndRedirectToLogin(ctx *gin.Context, returnURL string) {
	ctx.Redirect(http.StatusFound, "/login?next="+url.QueryEscape(returnURL))
}

func (c *OIDCController) redirectError(ctx *gin.Context, redirectURI, errCode, state string) {
	if redirectURI == "" {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(errors.New(errCode)))
		return
	}
	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("error", errCode)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	ctx.Redirect(http.StatusFound, u.String())
}

func (c *OIDCController) redirectSuccess(ctx *gin.Context, redirectURI, code, state string) {
	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	u.RawQuery = q.Encode()
	ctx.Redirect(http.StatusFound, u.String())
}

func (c *OIDCController) token(ctx *gin.Context) {
	var req oidc_dto.TokenRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, oidc_dto.NewOAuth2Error("invalid_request", ""))
		return
	}

	clientID := req.ClientID
	clientSecret := req.ClientSecret
	if clientID == "" {
		clientID, clientSecret, _ = ctx.Request.BasicAuth()
	}

	c.log.Debug("token request",
		"grant_type", req.GrantType,
		"client_id", clientID,
		"redirect_uri", req.RedirectURI,
		"code_len", len(req.Code),
		"refresh_token_len", len(req.RefreshToken),
	)

	if clientID == "" {
		ctx.JSON(http.StatusUnauthorized, oidc_dto.NewOAuth2Error("invalid_client", ""))
		return
	}

	client, err := c.service.ValidateClient(clientID, clientSecret, req.RedirectURI)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, oidc_dto.NewOAuth2Error("invalid_client", ""))
		return
	}

	var res *oidc_dto.TokenResponse
	if req.GrantType == "authorization_code" {
		res, err = c.service.ExchangeCode(ctx.Request.Context(), req, client)
	} else if req.GrantType == "refresh_token" {
		res, err = c.service.ExchangeRefreshToken(ctx.Request.Context(), req, client)
	} else {
		ctx.JSON(http.StatusBadRequest, oidc_dto.NewOAuth2Error("unsupported_grant_type", ""))
		return
	}

	if err != nil {
		c.log.Debug("token exchange failed", "error", err)
		ctx.JSON(http.StatusBadRequest, oidc_dto.NewOAuth2Error("invalid_grant", ""))
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *OIDCController) userinfo(ctx *gin.Context) {
	auth := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		ctx.JSON(http.StatusUnauthorized, oidc_dto.NewOAuth2Error("invalid_token", ""))
		return
	}
	tokenStr := strings.TrimPrefix(auth, "Bearer ")

	res, err := c.service.GetUserInfo(ctx.Request.Context(), tokenStr)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, oidc_dto.NewOAuth2Error("invalid_token", ""))
		return
	}
	c.log.Debug(
		"userinfo request",
		"user_id", res.Sub,
		"email", res.Email,
		"username", res.Username,
		"name", res.Name,
		"session_uuid", c.extractSessionUUIDFromToken(tokenStr))

	ctx.JSON(http.StatusOK, res)
}

func (c *OIDCController) revoke(ctx *gin.Context) {
	var req oidc_dto.RevokeRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, oidc_dto.NewOAuth2Error("invalid_request", ""))
		return
	}
	_ = c.service.Revoke(ctx.Request.Context(), req.Token)
	ctx.Status(http.StatusOK)
}

func (c *OIDCController) introspect(ctx *gin.Context) {
	var req oidc_dto.IntrospectRequest
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, oidc_dto.NewOAuth2Error("invalid_request", ""))
		return
	}
	res, err := c.service.Introspect(ctx.Request.Context(), req.Token)
	if err != nil {
		ctx.JSON(http.StatusOK, oidc_dto.IntrospectResponse{Active: false})
		return
	}
	ctx.JSON(http.StatusOK, res)
}

func (c *OIDCController) logout(ctx *gin.Context) {
	redirectURI := ctx.Query("post_logout_redirect_uri")
	idTokenHint := ctx.Query("id_token_hint")

	// Do NOT clear the IdP session yet — the user can choose to:
	//   • Log back into the app (SSO, no credentials needed — session still alive)
	//   • Sign out of min-idp (clears the session via /api/auth/logout)

	// Resolve SP display name from id_token_hint audience
	spName := "the application"
	if idTokenHint != "" {
		if claims, err := jwtutil.PayloadClaims(idTokenHint); err == nil {
			if aud, ok := claims["aud"].(string); ok && aud != "" {
				if client, err := c.service.spRepo.FindOIDCClientByClientID(aud); err == nil {
					if spEntity, err := c.service.spRepo.FindByID(client.SPID); err == nil {
						spName = spEntity.Name
					}
				}
			}
		}
	}

	views.LogoutTmpl.Render(ctx, views.LogoutViewModel{SPName: spName, ReturnURL: redirectURI})
}

func (c *OIDCController) extractSessionUUIDFromToken(tokenStr string) string {
	if tokenStr == "" {
		return ""
	}
	claims, err := jwtutil.PayloadClaims(tokenStr)
	if err != nil {
		return ""
	}
	sid, _ := claims["sid"].(string)
	return sid
}
