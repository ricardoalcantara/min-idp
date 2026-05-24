package authn

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	authn_dto "github.com/ricardoalcantara/min-idp/internal/authn/dto"
	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/jwtutil"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	"github.com/ricardoalcantara/min-idp/internal/rbac"
	"github.com/ricardoalcantara/min-idp/internal/session"
	"github.com/ricardoalcantara/min-idp/internal/views"
)

const LoginStateCookie = "min_idp_login_state"

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

func (c *AuthnController) landingPage(ctx *gin.Context) {
	views.LandingTmpl.Render(ctx, nil)
}

func (c *AuthnController) loginPage(ctx *gin.Context) {
	views.LoginTmpl.Render(ctx, map[string]string{"Next": ctx.Query("next"), "Error": ""})
}

func (c *AuthnController) infoPage(ctx *gin.Context) {
	cookie, err := ctx.Cookie(c.cfg.SessionCookie)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	rawJWT, err := c.cookieToken.Decode(cookie)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	claims, err := jwtutil.PayloadClaims(rawJWT)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/")
		return
	}
	// Convert roles []interface{} → []string for template range
	var roleList []string
	if raw, ok := claims["roles"].([]interface{}); ok {
		for _, r := range raw {
			if s, ok := r.(string); ok {
				roleList = append(roleList, s)
			}
		}
	}
	views.InfoTmpl.Render(ctx, map[string]interface{}{
		"Email":    claims["email"],
		"Name":     claims["name"],
		"UUID":     claims["sub"],
		"RoleList": roleList,
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

	u, err := c.service.Authenticate(input.Login, input.Password)
	if err != nil {
		if errors.Is(err, errInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, web.NewErrorDto(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}

	hasAPIUser, _ := c.rbacSvc.UserHasRole(u.ID, "api:user")
	hasAdmin, err := c.rbacSvc.UserHasRole(u.ID, "system:admin")
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

	token, err := c.mintToken(ctx, u.ID, u.UUID.String(), sess.UUID.String(), u.Email, u.Username, u.Name, sess.ExpiresAt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}

	ctx.JSON(http.StatusOK, authn_dto.LoginResponseDto{AccessToken: token})
}

func (c *AuthnController) loginForm(ctx *gin.Context) {
	login := ctx.PostForm("login")
	password := ctx.PostForm("password")
	next := ctx.PostForm("next")

	renderLoginError := func(status int, msg string) {
		ctx.Status(status)
		views.LoginTmpl.Render(ctx, map[string]string{"Next": next, "Error": msg})
	}

	u, err := c.service.Authenticate(login, password)
	if err != nil {
		if errors.Is(err, errInvalidCredentials) {
			renderLoginError(http.StatusUnauthorized, "Invalid credentials.")
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

	token, err := c.mintToken(ctx, u.ID, u.UUID.String(), sess.UUID.String(), u.Email, u.Username, u.Name, sess.ExpiresAt)
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
	if next := ctx.PostForm("next"); strings.HasPrefix(next, "/oauth2/") || strings.HasPrefix(next, "/saml/") {
		redirectURL = next
	}
	ctx.Redirect(http.StatusSeeOther, redirectURL)
}

func (c *AuthnController) logout(ctx *gin.Context) {
	sid := c.extractSessionUUID(ctx)
	if sid != "" {
		_ = c.sessionSvc.Revoke(ctx.Request.Context(), sid)
	}
	ctx.SetCookie(c.cfg.SessionCookie, "", -1, "/", "", false, true)

	// If a redirect target is provided (e.g. from a browser form), honour it.
	// API clients that don't pass ?redirect= still get 204.
	if redirect := ctx.Query("redirect"); redirect != "" {
		ctx.Redirect(http.StatusSeeOther, redirect)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *AuthnController) register(ctx *gin.Context) {
	if !c.cfg.FeatureAPIRegistration {
		ctx.JSON(http.StatusForbidden, web.NewErrorDto(errors.New("registration is disabled")))
		return
	}
	ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented"))
}

const forgotPasswordGenericMsg = "If that email is registered, a reset link has been sent."

func (c *AuthnController) forgotPasswordPage(ctx *gin.Context) {
	views.ForgotPasswordTmpl.Render(ctx, map[string]string{"Error": "", "Message": ""})
}

func (c *AuthnController) forgotPassword(ctx *gin.Context) {
	ct := ctx.GetHeader("Content-Type")
	isForm := strings.HasPrefix(ct, "application/x-www-form-urlencoded")

	var input authn_dto.ForgotPasswordDto
	var bindErr error
	if isForm {
		bindErr = ctx.ShouldBind(&input)
	} else {
		bindErr = ctx.ShouldBindJSON(&input)
	}
	if bindErr != nil {
		if isForm {
			ctx.Status(http.StatusBadRequest)
			views.ForgotPasswordTmpl.Render(ctx, map[string]string{"Error": "Please enter a valid email.", "Message": ""})
			return
		}
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(bindErr))
		return
	}

	// Errors here (rejected method, infra failures) are not surfaced to the
	// caller — the response is the same generic 200 either way. The service
	// logs/propagates as needed.
	_ = c.service.RequestPasswordReset(ctx.Request.Context(), input.Email, input.CodeChallenge, input.CodeChallengeMethod)

	if isForm {
		views.ForgotPasswordTmpl.Render(ctx, map[string]string{"Error": "", "Message": forgotPasswordGenericMsg})
		return
	}
	ctx.JSON(http.StatusOK, web.NewMessageDto(forgotPasswordGenericMsg))
}

func (c *AuthnController) resetPasswordPage(ctx *gin.Context) {
	views.ResetPasswordTmpl.Render(ctx, map[string]string{
		"Token":   ctx.Query("token"),
		"Error":   "",
		"Message": "",
	})
}

func (c *AuthnController) resetPassword(ctx *gin.Context) {
	ct := ctx.GetHeader("Content-Type")
	isForm := strings.HasPrefix(ct, "application/x-www-form-urlencoded")

	var input authn_dto.ResetPasswordDto
	var bindErr error
	if isForm {
		bindErr = ctx.ShouldBind(&input)
	} else {
		bindErr = ctx.ShouldBindJSON(&input)
	}
	if bindErr != nil {
		if isForm {
			ctx.Status(http.StatusBadRequest)
			views.ResetPasswordTmpl.Render(ctx, map[string]string{
				"Token":   input.Token,
				"Error":   "Please enter a password of at least 8 characters.",
				"Message": "",
			})
			return
		}
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(bindErr))
		return
	}

	if err := c.service.ResetPassword(ctx.Request.Context(), input.Token, input.Password, input.CodeVerifier); err != nil {
		if errors.Is(err, errInvalidResetToken) || errors.Is(err, errWeakPassword) {
			if isForm {
				ctx.Status(http.StatusBadRequest)
				views.ResetPasswordTmpl.Render(ctx, map[string]string{
					"Token":   input.Token,
					"Error":   "Reset link is invalid or has expired. Please request a new one.",
					"Message": "",
				})
				return
			}
			ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
			return
		}
		if isForm {
			ctx.Status(http.StatusInternalServerError)
			views.ResetPasswordTmpl.Render(ctx, map[string]string{
				"Token":   input.Token,
				"Error":   "Internal error. Please try again.",
				"Message": "",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}

	if isForm {
		views.ResetPasswordTmpl.Render(ctx, map[string]string{
			"Token":   "",
			"Error":   "",
			"Message": "Your password has been reset. You can now sign in.",
		})
		return
	}
	ctx.JSON(http.StatusOK, web.NewMessageDto("password updated"))
}

func (c *AuthnController) mintToken(ctx *gin.Context, userID uint, userUUID, sessionUUID, email, username, name string, expiresAt time.Time) (string, error) {
	roles, _ := c.rbacSvc.GetUserRoleNames(userID)

	key, meta, err := c.ks.ActivePrivateKey(ctx.Request.Context(), keystore_entities.ProtocolOIDC)
	if err != nil {
		return "", err
	}

	return MintSessionJWT(key, SigningConfig{KID: meta.KID, Issuer: c.cfg.ExternalURL},
		userID, userUUID, sessionUUID, email, username, name, roles, time.Until(expiresAt))
}

func (c *AuthnController) extractSessionUUID(ctx *gin.Context) string {
	if cookie, err := ctx.Cookie(c.cfg.SessionCookie); err == nil {
		if rawJWT, err := c.cookieToken.Decode(cookie); err == nil {
			if claims, err := jwtutil.PayloadClaims(rawJWT); err == nil {
				if sid, ok := claims["sid"].(string); ok {
					return sid
				}
			}
		}
	}
	if auth := ctx.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		if claims, err := jwtutil.PayloadClaims(token); err == nil {
			if sid, ok := claims["sid"].(string); ok {
				return sid
			}
		}
	}
	return ""
}

