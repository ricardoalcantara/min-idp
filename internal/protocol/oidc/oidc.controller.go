package oidc

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/config"
	localcrypto "github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	oidc_dto "github.com/ricardoalcantara/min-idp/internal/protocol/oidc/dto"
	"github.com/ricardoalcantara/min-idp/internal/session"
	"github.com/ricardoalcantara/min-idp/internal/sp"
)

var logoutTmpl = template.Must(template.New("logout").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Signed out — min-idp</title>
  <style>
    :root {
      --bg:#f5f7fa;--card:#fff;--shadow:0 4px 24px rgba(0,0,0,.08);
      --text:#111827;--sub:#6b7280;--border:#e5e7eb;--toggle-bg:#e5e7eb;
    }
    :root:has(#dark-toggle:checked){
      --bg:#0f1117;--card:#1a1d27;--shadow:0 4px 24px rgba(0,0,0,.4);
      --text:#f3f4f6;--sub:#9ca3af;--border:#2d3149;--toggle-bg:#374151;
    }
    *,*::before,*::after{box-sizing:border-box;margin:0;padding:0;}
    body{min-height:100vh;display:flex;align-items:center;justify-content:center;
      background:var(--bg);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;
      color:var(--text);transition:background .25s,color .25s;}
    #dark-toggle{display:none;}
    .toggle-btn{position:fixed;top:16px;right:16px;width:38px;height:38px;
      background:var(--toggle-bg);border-radius:50%;cursor:pointer;
      display:flex;align-items:center;justify-content:center;font-size:18px;
      transition:background .25s;user-select:none;}
    .toggle-btn::after{content:var(--toggle-ico,"🌙");}
    :root:has(#dark-toggle:checked) .toggle-btn::after{content:"☀️";}
    .card{background:var(--card);border-radius:14px;box-shadow:var(--shadow);
      padding:40px 36px;width:100%;max-width:400px;text-align:center;
      transition:background .25s,box-shadow .25s;}
    .logo{width:44px;height:44px;background:linear-gradient(135deg,#4f46e5,#7c3aed);
      border-radius:10px;display:flex;align-items:center;justify-content:center;
      color:#fff;font-weight:700;font-size:18px;margin:0 auto 20px;}
    .check{width:48px;height:48px;background:#f0fdf4;border-radius:50%;
      display:flex;align-items:center;justify-content:center;margin:0 auto 16px;font-size:22px;}
    h1{font-size:20px;font-weight:600;margin-bottom:8px;}
    .sub{font-size:14px;color:var(--sub);margin-bottom:28px;line-height:1.5;}
    .divider{height:1px;background:var(--border);margin:0 -36px 24px;}
    .btn{display:block;width:100%;padding:11px;border-radius:8px;font-size:14px;
      font-weight:600;cursor:pointer;border:none;text-decoration:none;
      transition:opacity .2s;margin-bottom:10px;}
    .btn:hover{opacity:.85;}
    .btn-primary{background:#4f46e5;color:#fff;}
    .btn-secondary{background:transparent;color:var(--sub);border:1.5px solid var(--border);}
  </style>
</head>
<body>
  <input type="checkbox" id="dark-toggle" onchange="localStorage.setItem('theme',this.checked?'dark':'light')">
  <label class="toggle-btn" for="dark-toggle"></label>
  <script>
    (function(){
      var el=document.getElementById('dark-toggle');
      var s=localStorage.getItem('theme');
      var d=s?s==='dark':matchMedia('(prefers-color-scheme:dark)').matches;
      if(d)el.checked=true;
    })();
  </script>
  <div class="card">
    <div class="logo">M</div>
    <div class="check">✓</div>
    <h1>Signed out of {{.SPName}}</h1>
    <p class="sub">Your session with <strong>{{.SPName}}</strong> has ended.</p>
    <div class="divider"></div>
    {{if .ReturnURL}}
    <a class="btn btn-primary" href="{{.ReturnURL}}">Log back into {{.SPName}}</a>
    {{end}}
    <form method="POST" action="/api/auth/logout?redirect=/info" style="margin:0">
      <button class="btn btn-secondary" type="submit">Sign out of min-idp</button>
    </form>
  </div>
</body>
</html>`))

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
		if claims, err := jwtPayloadClaims(idTokenHint); err == nil {
			if aud, ok := claims["aud"].(string); ok && aud != "" {
				if client, err := c.service.spRepo.FindOIDCClientByClientID(aud); err == nil {
					if spEntity, err := c.service.spRepo.FindByID(client.SPID); err == nil {
						spName = spEntity.Name
					}
				}
			}
		}
	}

	ctx.Header("Content-Type", "text/html; charset=utf-8")
	_ = logoutTmpl.Execute(ctx.Writer, map[string]any{
		"SPName":    spName,
		"ReturnURL": redirectURI,
	})
}

func (c *OIDCController) extractSessionUUIDFromToken(tokenStr string) string {
	if tokenStr == "" {
		return ""
	}
	claims, err := jwtPayloadClaims(tokenStr)
	if err != nil {
		return ""
	}
	sid, _ := claims["sid"].(string)
	return sid
}

// jwtPayloadClaims decodes the JWT payload without verifying the signature.
func jwtPayloadClaims(tokenStr string) (map[string]any, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid jwt")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}
