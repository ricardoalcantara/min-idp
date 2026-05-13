package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-minstack/core"
	mgin "github.com/go-minstack/gin"
	"github.com/go-minstack/migration"
	"github.com/ricardoalcantara/min-idp/internal/audit"
	"github.com/ricardoalcantara/min-idp/internal/authn"
	"github.com/ricardoalcantara/min-idp/internal/bootstrap"
	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	"github.com/ricardoalcantara/min-idp/internal/protocol/oidc"
	"github.com/ricardoalcantara/min-idp/internal/protocol/saml"
	"github.com/ricardoalcantara/min-idp/internal/rbac"
	"github.com/ricardoalcantara/min-idp/internal/session"
	"github.com/ricardoalcantara/min-idp/internal/sp"
	"github.com/ricardoalcantara/min-idp/internal/storage"
	"github.com/ricardoalcantara/min-idp/internal/users"
	"github.com/ricardoalcantara/min-idp/migrations"
)

func registerAdminRoutes(
	r *gin.Engine,
	cfg *config.Config,
	sessionSvc *session.SessionService,
	rbacSvc *rbac.RBACService,
	ks keystore.KeyStore,
	kv kvstore.KVStore,
	cookieToken *session.CookieTokenService,
	userCtrl *users.UserController,
	rbacCtrl *rbac.RBACController,
	groupCtrl *rbac.GroupController,
	spCtrl *sp.SPController,
	keyCtrl *keystore.KeyStoreController,
	auditCtrl *audit.AuditController,
) {
	adminMW := []gin.HandlerFunc{
		sessionSvc.APIMiddleware(cfg.SessionCookie, ks, kv, cookieToken),
		session.RequireSession(),
		rbac.RequirePermission(rbacSvc, "system:admin"),
	}
	users.RegisterUserRoutes(r, userCtrl, adminMW...)
	rbac.RegisterRoutes(r, rbacCtrl, adminMW...)
	rbac.RegisterGroupRoutes(r, groupCtrl, adminMW...)
	sp.RegisterRoutes(r, spCtrl, adminMW...)
	keystore.RegisterRoutes(r, keyCtrl, adminMW...)
	audit.RegisterRoutes(r, auditCtrl, adminMW...)
}

func main() {
	app := core.New(
		config.Module(),
		mgin.Module(),
		storage.Module(),
		kvstore.Module(),
		migration.Module(migrations.FS),
	)

	users.Register(app)
	session.Register(app)
	rbac.Register(app)
	sp.Register(app)
	keystore.Register(app)
	audit.Register(app)
	authn.Register(app)
	bootstrap.Register(app)
	oidc.Register(app)
	saml.Register(app)

	app.Invoke(migration.Run)
	app.Invoke(users.RegisterMeRoutes)
	app.Invoke(registerAdminRoutes)
	app.Run()
}
