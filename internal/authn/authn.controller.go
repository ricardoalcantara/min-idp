package authn

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	authn_dto "github.com/ricardoalcantara/min-idp/internal/authn/dto"
	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	"github.com/ricardoalcantara/min-idp/internal/rbac"
	"github.com/ricardoalcantara/min-idp/internal/session"
)

const LoginStateCookie = "min_idp_login_state"

var loginTmpl = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Sign in</title></head>
<body>
  <form method="POST" action="/api/auth/login">
    <input type="hidden" name="state" value="{{.State}}">
    <label>Email<br><input type="email" name="email" required autofocus></label><br>
    <label>Password<br><input type="password" name="password" required></label><br>
    {{if .Error}}<p style="color:red">{{.Error}}</p>{{end}}
    <button type="submit">Sign in</button>
  </form>
</body>
</html>`))

var infoTmpl = template.Must(template.New("info").Parse(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>min-idp</title></head>
<body>
  {{if .Email}}
    <h2>Signed in</h2>
    <p><strong>Email:</strong> {{.Email}}</p>
    <p><strong>ID:</strong> {{.UUID}}</p>
    <p><strong>Roles:</strong> {{.Roles}}</p>
    <form method="POST" action="/api/auth/logout"><button type="submit">Sign out</button></form>
  {{else}}
    <h2>Not signed in</h2>
    <a href="/login">Sign in</a>
  {{end}}
</body>
</html>`))

type AuthnController struct {
	service     *AuthnService
	sessionSvc  *session.SessionService
	rbacSvc     *rbac.RBACService
	ks          keystore.KeyStore
	kv          kvstore.KVStore
	cookieToken *session.CookieTokenService
	cfg         *config.Config
}

func NewAuthnController(
	service *AuthnService,
	sessionSvc *session.SessionService,
	rbacSvc *rbac.RBACService,
	ks keystore.KeyStore,
	kv kvstore.KVStore,
	cookieToken *session.CookieTokenService,
	cfg *config.Config,
) *AuthnController {
	return &AuthnController{
		service:     service,
		sessionSvc:  sessionSvc,
		rbacSvc:     rbacSvc,
		ks:          ks,
		kv:          kv,
		cookieToken: cookieToken,
		cfg:         cfg,
	}
}

func (c *AuthnController) loginPage(ctx *gin.Context) {
	state, _ := ctx.Cookie(LoginStateCookie)
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	_ = loginTmpl.Execute(ctx.Writer, map[string]string{"State": state, "Error": ""})
}

func (c *AuthnController) infoPage(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	cookie, err := ctx.Cookie(c.cfg.SessionCookie)
	if err != nil {
		_ = infoTmpl.Execute(ctx.Writer, map[string]interface{}{"Email": ""})
		return
	}
	rawJWT, err := c.cookieToken.Decode(cookie)
	if err != nil {
		_ = infoTmpl.Execute(ctx.Writer, map[string]interface{}{"Email": ""})
		return
	}
	claims, err := jwtPayloadClaims(rawJWT)
	if err != nil {
		_ = infoTmpl.Execute(ctx.Writer, map[string]interface{}{"Email": ""})
		return
	}
	_ = infoTmpl.Execute(ctx.Writer, map[string]interface{}{
		"Email": claims["email"],
		"UUID":  claims["sub"],
		"Roles": claims["roles"],
	})
}

func (c *AuthnController) login(ctx *gin.Context) {
	ct := ctx.GetHeader("Content-Type")
	if strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		c.loginForm(ctx)
		return
	}

	var input authn_dto.LoginDto
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}

	u, err := c.service.Authenticate(input.Email, input.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}

	hasAPIUser, _ := c.rbacSvc.UserHasPermission(u.ID, "api:user")
	hasAdmin, err := c.rbacSvc.UserHasPermission(u.ID, "system:admin")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	if !hasAPIUser && !hasAdmin {
		ctx.JSON(http.StatusForbidden, web.NewErrorDto(errors.New("api login not permitted for this account")))
		return
	}

	sess, err := c.sessionSvc.Create(ctx.Request.Context(), u.ID, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}

	token, err := c.mintToken(ctx, u.ID, u.UUID.String(), sess.UUID.String(), u.Email, sess.ExpiresAt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}

	ctx.JSON(http.StatusOK, authn_dto.LoginResponseDto{AccessToken: token})
}

func (c *AuthnController) loginForm(ctx *gin.Context) {
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")
	state := ctx.PostForm("state")

	renderLoginError := func(status int, msg string) {
		ctx.Status(status)
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		_ = loginTmpl.Execute(ctx.Writer, map[string]string{"State": state, "Error": msg})
	}

	u, err := c.service.Authenticate(email, password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			renderLoginError(http.StatusUnauthorized, "Invalid email or password.")
		} else {
			renderLoginError(http.StatusInternalServerError, "Internal error. Please try again.")
		}
		return
	}

	sess, err := c.sessionSvc.Create(ctx.Request.Context(), u.ID, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		renderLoginError(http.StatusInternalServerError, "Internal error. Please try again.")
		return
	}

	token, err := c.mintToken(ctx, u.ID, u.UUID.String(), sess.UUID.String(), u.Email, sess.ExpiresAt)
	if err != nil {
		renderLoginError(http.StatusInternalServerError, "Internal error. Please try again.")
		return
	}

	encrypted, err := c.cookieToken.Encode(token)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.SetCookie(c.cfg.SessionCookie, encrypted, int(c.cfg.SessionTTL/time.Second), "/", "", ctx.Request.TLS != nil, true)

	redirectURL := "/info"
	if state != "" {
		if val, err := c.kv.Get(ctx.Request.Context(), "login_state:"+state); err == nil && len(val) > 0 {
			redirectURL = string(val)
			_ = c.kv.Delete(ctx.Request.Context(), "login_state:"+state)
		}
	}
	ctx.Redirect(http.StatusSeeOther, redirectURL)
}

func (c *AuthnController) logout(ctx *gin.Context) {
	sid := c.extractSessionUUID(ctx)
	if sid != "" {
		_ = c.sessionSvc.Revoke(ctx.Request.Context(), sid)
	}
	ctx.SetCookie(c.cfg.SessionCookie, "", -1, "/", "", false, true)
	ctx.Status(http.StatusNoContent)
}

func (c *AuthnController) register(ctx *gin.Context) {
	if !c.cfg.FeatureAPIRegistration {
		ctx.JSON(http.StatusForbidden, web.NewErrorDto(errors.New("registration is disabled")))
		return
	}
	ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented"))
}

func (c *AuthnController) mintToken(ctx *gin.Context, userID uint, userUUID, sessionUUID, email string, expiresAt time.Time) (string, error) {
	roles, _ := c.rbacSvc.GetUserPermissions(userID)

	key, meta, err := c.ks.ActivePrivateKey(ctx.Request.Context(), keystore_entities.ProtocolOIDC)
	if err != nil {
		return "", err
	}

	return MintSessionJWT(key, meta.KID, meta.Algorithm, c.cfg.ExternalURL,
		userID, userUUID, sessionUUID, email, roles, time.Until(expiresAt))
}

func (c *AuthnController) extractSessionUUID(ctx *gin.Context) string {
	if cookie, err := ctx.Cookie(c.cfg.SessionCookie); err == nil {
		if rawJWT, err := c.cookieToken.Decode(cookie); err == nil {
			if claims, err := jwtPayloadClaims(rawJWT); err == nil {
				if sid, ok := claims["sid"].(string); ok {
					return sid
				}
			}
		}
	}
	if auth := ctx.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		if claims, err := jwtPayloadClaims(token); err == nil {
			if sid, ok := claims["sid"].(string); ok {
				return sid
			}
		}
	}
	return ""
}

// jwtPayloadClaims decodes the JWT payload without verifying the signature.
// Used only for display (infoPage) and logout SID extraction —
// the middleware already verified the signature before the handler runs.
func jwtPayloadClaims(tokenStr string) (map[string]interface{}, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid jwt format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}
