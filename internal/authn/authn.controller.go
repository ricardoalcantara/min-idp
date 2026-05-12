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
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Sign in — min-idp</title>
  <style>
    :root {
      --bg:         #f5f7fa;
      --card:       #ffffff;
      --shadow:     0 4px 24px rgba(0,0,0,.08);
      --text:       #111827;
      --sub:        #6b7280;
      --label:      #374151;
      --border:     #d1d5db;
      --input-bg:   #ffffff;
      --input-text: #111827;
      --error-bg:   #fef2f2;
      --error-bd:   #fecaca;
      --error-tx:   #dc2626;
      --toggle-bg:  #e5e7eb;
      --toggle-ico: "🌙";
    }
    :root:has(#dark-toggle:checked) {
      --bg:         #0f1117;
      --card:       #1a1d27;
      --shadow:     0 4px 24px rgba(0,0,0,.4);
      --text:       #f3f4f6;
      --sub:        #9ca3af;
      --label:      #d1d5db;
      --border:     #374151;
      --input-bg:   #111827;
      --input-text: #f3f4f6;
      --error-bg:   #2d1515;
      --error-bd:   #7f1d1d;
      --error-tx:   #fca5a5;
      --toggle-bg:  #374151;
      --toggle-ico: "☀️";
    }
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    body {
      min-height: 100vh;
      display: flex; align-items: center; justify-content: center;
      background: var(--bg);
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      color: var(--text);
      transition: background .25s, color .25s;
    }
    #dark-toggle { display: none; }
    .toggle-btn {
      position: fixed; top: 16px; right: 16px;
      width: 38px; height: 38px;
      background: var(--toggle-bg);
      border-radius: 50%; cursor: pointer;
      display: flex; align-items: center; justify-content: center;
      font-size: 18px; transition: background .25s;
      user-select: none;
    }
    .toggle-btn::after { content: var(--toggle-ico); }
    .card {
      background: var(--card);
      border-radius: 14px;
      box-shadow: var(--shadow);
      padding: 40px 36px;
      width: 100%; max-width: 400px;
      transition: background .25s, box-shadow .25s;
    }
    .logo {
      width: 44px; height: 44px;
      background: linear-gradient(135deg, #4f46e5, #7c3aed);
      border-radius: 10px;
      display: flex; align-items: center; justify-content: center;
      color: #fff; font-weight: 700; font-size: 18px;
      margin: 0 auto 20px;
    }
    h1 { text-align: center; font-size: 22px; font-weight: 600; margin-bottom: 6px; }
    .sub { text-align: center; font-size: 14px; color: var(--sub); margin-bottom: 28px; }
    .field-label {
      display: block; font-size: 13px; font-weight: 500;
      color: var(--label); margin-bottom: 6px;
    }
    input[type=email], input[type=password] {
      width: 100%; padding: 10px 12px;
      background: var(--input-bg); color: var(--input-text);
      border: 1.5px solid var(--border); border-radius: 8px;
      font-size: 15px; outline: none;
      transition: border-color .2s, background .25s, color .25s;
      margin-bottom: 16px;
    }
    input[type=email]:focus, input[type=password]:focus { border-color: #4f46e5; }
    .error {
      background: var(--error-bg); border: 1px solid var(--error-bd);
      color: var(--error-tx); border-radius: 8px;
      padding: 10px 12px; font-size: 13px; margin-bottom: 16px;
    }
    .submit-btn {
      width: 100%; padding: 11px;
      background: #4f46e5; color: #fff;
      border: none; border-radius: 8px;
      font-size: 15px; font-weight: 600; cursor: pointer;
      transition: background .2s;
    }
    .submit-btn:hover { background: #4338ca; }
  </style>
</head>
<body>
  <input type="checkbox" id="dark-toggle" onchange="localStorage.setItem('theme',this.checked?'dark':'light')">
  <label class="toggle-btn" for="dark-toggle"></label>
  <script>
    (function(){
      var s=localStorage.getItem('theme');
      var d=s?s==='dark':window.matchMedia('(prefers-color-scheme:dark)').matches;
      if(d)document.getElementById('dark-toggle').checked=true;
    })();
  </script>
  <div class="card">
    <div class="logo">M</div>
    <h1>Welcome back</h1>
    <p class="sub">Sign in to continue</p>
    {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
    <form method="POST" action="/api/auth/login">
      <input type="hidden" name="next" value="{{.Next}}">
      <label class="field-label" for="email">Email</label>
      <input id="email" type="email" name="email" placeholder="you@example.com" required autofocus>
      <label class="field-label" for="password">Password</label>
      <input id="password" type="password" name="password" placeholder="••••••••" required>
      <button class="submit-btn" type="submit">Sign in</button>
    </form>
  </div>
</body>
</html>`))

var infoTmpl = template.Must(template.New("info").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>min-idp</title>
  <style>
    :root {
      --bg:#f5f7fa;--card:#fff;--shadow:0 4px 24px rgba(0,0,0,.08);
      --text:#111827;--sub:#6b7280;--border:#e5e7eb;
      --card-header:#f9fafb;--muted:#6b7280;
      --attr-bg:rgba(238,242,255,.5);--attr-text:#4338ca;
      --toggle-bg:#e5e7eb;
    }
    :root:has(#dark-toggle:checked){
      --bg:#0f1117;--card:#1a1d27;--shadow:0 4px 24px rgba(0,0,0,.4);
      --text:#f3f4f6;--sub:#9ca3af;--border:#2d3149;
      --card-header:#1e2235;--muted:#9ca3af;
      --attr-bg:rgba(49,46,129,.2);--attr-text:#a5b4fc;
      --toggle-bg:#374151;
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
    .wrap{width:100%;max-width:480px;padding:1rem;}
    .header{text-align:center;margin-bottom:2rem;}
    .logo{width:44px;height:44px;background:linear-gradient(135deg,#4f46e5,#7c3aed);
      border-radius:10px;display:flex;align-items:center;justify-content:center;
      color:#fff;font-weight:700;font-size:18px;margin:0 auto 16px;}
    h1{font-size:22px;font-weight:600;margin-bottom:4px;}
    .sub{font-size:14px;color:var(--muted);}
    .card{background:var(--card);border:1px solid var(--border);border-radius:14px;
      box-shadow:var(--shadow);overflow:hidden;margin-bottom:12px;}
    .card-header{background:var(--card-header);border-bottom:1px solid var(--border);
      padding:0.6rem 1.25rem;font-size:0.7rem;font-weight:600;text-transform:uppercase;
      letter-spacing:.05em;color:var(--muted);}
    .card-body{padding:1rem 1.25rem;}
    .field{margin-bottom:10px;}
    .field:last-child{margin-bottom:0;}
    .field-label{font-size:11px;font-weight:600;text-transform:uppercase;
      letter-spacing:.05em;color:var(--muted);margin-bottom:3px;}
    .field-value{font-size:14px;font-family:monospace;word-break:break-all;
      color:var(--attr-text);}
    .roles{display:flex;flex-wrap:wrap;gap:6px;margin-top:4px;}
    .role{font-size:12px;padding:2px 8px;border-radius:99px;
      background:var(--attr-bg);color:var(--attr-text);font-weight:500;}
    .status{background:var(--card);border:1px solid var(--border);border-radius:14px;
      padding:1rem 1.25rem;display:flex;align-items:center;justify-content:space-between;
      box-shadow:var(--shadow);margin-bottom:12px;}
    .dot{width:10px;height:10px;border-radius:50%;background:#10b981;
      box-shadow:0 0 0 4px rgba(16,185,129,.15);margin-right:12px;flex-shrink:0;}
    .btn-logout{padding:8px 16px;background:var(--border);color:var(--text);
      border:none;border-radius:8px;font-size:13px;font-weight:500;cursor:pointer;
      transition:opacity .2s;}
    .btn-logout:hover{opacity:.75;}
    .sign-in-card{background:var(--card);border:1px solid var(--border);border-radius:14px;
      box-shadow:var(--shadow);padding:2rem;text-align:center;}
    .sign-in-card p{color:var(--muted);font-size:14px;margin-bottom:1.25rem;}
    .btn-login{display:inline-flex;align-items:center;gap:8px;padding:10px 24px;
      background:#4f46e5;color:#fff;border:none;border-radius:8px;
      font-size:14px;font-weight:600;cursor:pointer;text-decoration:none;
      transition:background .2s;}
    .btn-login:hover{background:#4338ca;}
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
  <div class="wrap">
    <div class="header">
      <div class="logo">M</div>
      <h1>min-idp</h1>
      <p class="sub">Identity Provider</p>
    </div>
    {{if .Email}}
      <div class="status">
        <div style="display:flex;align-items:center">
          <span class="dot"></span>
          <div>
            <div style="font-weight:600;font-size:14px">Signed in</div>
            <div style="font-size:12px;color:var(--muted)">{{.Email}}</div>
          </div>
        </div>
        <form method="POST" action="/api/auth/logout?redirect=/info" style="margin:0">
          <button class="btn-logout" type="submit">Sign out</button>
        </form>
      </div>
      <div class="card">
        <div class="card-header">Session</div>
        <div class="card-body">
          <div class="field">
            <div class="field-label">UUID</div>
            <div class="field-value">{{.UUID}}</div>
          </div>
          <div class="field">
            <div class="field-label">Email</div>
            <div class="field-value">{{.Email}}</div>
          </div>
          <div class="field">
            <div class="field-label">Roles</div>
            <div class="roles">{{range .RoleList}}<span class="role">{{.}}</span>{{end}}</div>
          </div>
        </div>
      </div>
    {{else}}
      <div class="sign-in-card">
        <p>You are not signed in to min-idp.</p>
        <a class="btn-login" href="/login">
          <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
              d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"/>
          </svg>
          Sign in
        </a>
      </div>
    {{end}}
  </div>
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
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	_ = loginTmpl.Execute(ctx.Writer, map[string]string{"Next": ctx.Query("next"), "Error": ""})
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
	// Convert roles []interface{} → []string for template range
	var roleList []string
	if raw, ok := claims["roles"].([]interface{}); ok {
		for _, r := range raw {
			if s, ok := r.(string); ok {
				roleList = append(roleList, s)
			}
		}
	}
	_ = infoTmpl.Execute(ctx.Writer, map[string]interface{}{
		"Email":    claims["email"],
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
	next := ctx.PostForm("next")

	renderLoginError := func(status int, msg string) {
		ctx.Status(status)
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		_ = loginTmpl.Execute(ctx.Writer, map[string]string{"Next": next, "Error": msg})
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

func (c *AuthnController) mintToken(ctx *gin.Context, userID uint, userUUID, sessionUUID, email string, expiresAt time.Time) (string, error) {
	roles, _ := c.rbacSvc.GetUserPermissions(userID)

	key, meta, err := c.ks.ActivePrivateKey(ctx.Request.Context(), keystore_entities.ProtocolOIDC)
	if err != nil {
		return "", err
	}

	return MintSessionJWT(key, SigningConfig{KID: meta.KID, Issuer: c.cfg.ExternalURL},
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
