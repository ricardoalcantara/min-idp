package authn

import (
	"errors"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	authn_dto "github.com/ricardoalcantara/min-idp/internal/authn/dto"
	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	"github.com/ricardoalcantara/min-idp/internal/session"
)

// LoginStateCookie is the cookie name used by OIDC/SAML authorize flows to
// stash a pending request key before redirecting to /login.
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

type AuthnController struct {
	service    *AuthnService
	sessionSvc *session.SessionService
	kv         kvstore.KVStore
	cfg        *config.Config
}

func NewAuthnController(service *AuthnService, sessionSvc *session.SessionService, kv kvstore.KVStore, cfg *config.Config) *AuthnController {
	return &AuthnController{service: service, sessionSvc: sessionSvc, kv: kv, cfg: cfg}
}

func (c *AuthnController) loginPage(ctx *gin.Context) {
	state, _ := ctx.Cookie(LoginStateCookie)
	_ = loginTmpl.Execute(ctx.Writer, map[string]string{"State": state, "Error": ""})
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

	sess, err := c.sessionSvc.Create(ctx.Request.Context(), u.ID, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}

	ctx.SetCookie(c.cfg.SessionCookie, sess.UUID.String(), int(c.cfg.SessionTTL/time.Second), "/", "", ctx.Request.TLS != nil, true)
	ctx.JSON(http.StatusOK, authn_dto.LoginResponseDto{SessionID: sess.UUID})
}

func (c *AuthnController) loginForm(ctx *gin.Context) {
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")
	state := ctx.PostForm("state")

	u, err := c.service.Authenticate(email, password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			ctx.Status(http.StatusUnauthorized)
			_ = loginTmpl.Execute(ctx.Writer, map[string]string{"State": state, "Error": "Invalid email or password."})
		} else {
			ctx.Status(http.StatusInternalServerError)
			_ = loginTmpl.Execute(ctx.Writer, map[string]string{"State": state, "Error": "Internal error. Please try again."})
		}
		return
	}

	sess, err := c.sessionSvc.Create(ctx.Request.Context(), u.ID, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		_ = loginTmpl.Execute(ctx.Writer, map[string]string{"State": state, "Error": "Internal error. Please try again."})
		return
	}

	ctx.SetCookie(c.cfg.SessionCookie, sess.UUID.String(), int(c.cfg.SessionTTL/time.Second), "/", "", ctx.Request.TLS != nil, true)

	redirectURL := c.cfg.PostLoginURL
	if state != "" {
		if val, err := c.kv.Get(ctx.Request.Context(), "login_state:"+state); err == nil && len(val) > 0 {
			redirectURL = string(val)
			_ = c.kv.Delete(ctx.Request.Context(), "login_state:"+state)
		}
	}

	ctx.Redirect(http.StatusSeeOther, redirectURL)
}

func (c *AuthnController) logout(ctx *gin.Context) {
	// Support both cookie (browser) and bearer (API client) revocation.
	token := ""
	if cookie, err := ctx.Cookie(c.cfg.SessionCookie); err == nil {
		token = cookie
	} else if auth := ctx.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		token = strings.TrimPrefix(auth, "Bearer ")
	}
	if token != "" {
		if sess, err := c.sessionSvc.GetByUUID(ctx.Request.Context(), token); err == nil {
			_ = c.sessionSvc.Revoke(ctx.Request.Context(), sess.ID)
		}
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
