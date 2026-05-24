package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

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
	"github.com/ricardoalcantara/min-idp/internal/notification"
	"github.com/ricardoalcantara/min-idp/internal/protocol/oidc"
	"github.com/ricardoalcantara/min-idp/internal/protocol/saml"
	"github.com/ricardoalcantara/min-idp/internal/rbac"
	"github.com/ricardoalcantara/min-idp/internal/session"
	"github.com/ricardoalcantara/min-idp/internal/sp"
	"github.com/ricardoalcantara/min-idp/internal/storage"
	"github.com/ricardoalcantara/min-idp/internal/users"
	"github.com/ricardoalcantara/min-idp/migrations"
	"github.com/stretchr/testify/require"
)

const (
	testMasterKey  = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	testAdminEmail = "admin@min-idp.local"
	testAdminPass  = "e2e-test-password"
)

type testApp struct {
	engine *gin.Engine
	cfg    *config.Config
}

func setupApp(t *testing.T) *testApp {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")

	t.Setenv("MIN_IDP_EXTERNAL_URL", "http://localhost:9999")
	t.Setenv("MIN_IDP_DB_DRIVER", "sqlite")
	t.Setenv("MIN_IDP_DB_DSN", dbPath)
	t.Setenv("MIN_IDP_MASTER_KEY", testMasterKey)
	t.Setenv("MIN_IDP_ADMIN_PASSWORD", testAdminPass)
	t.Setenv("MINSTACK_PORT", "0")

	gin.SetMode(gin.TestMode)

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
	notification.Register(app)
	authn.Register(app)
	bootstrap.Register(app)
	oidc.Register(app)
	saml.Register(app)

	app.Invoke(migration.Run)
	app.Invoke(users.RegisterMeRoutes)

	var engine *gin.Engine
	var cfg *config.Config
	app.Invoke(func(e *gin.Engine, c *config.Config) {
		engine = e
		cfg = c
	})

	ctx := context.Background()
	require.NoError(t, app.Start(ctx))
	t.Cleanup(func() { _ = app.Stop(ctx) })

	return &testApp{engine: engine, cfg: cfg}
}

// request sends an HTTP request to the test engine.
// Pass "Bearer <token>" or a cookie header via the headers map.
func (a *testApp) request(t *testing.T, method, path string, body any, headers ...map[string]string) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		require.NoError(t, json.NewEncoder(&buf).Encode(body))
	}

	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, h := range headers {
		for k, v := range h {
			req.Header.Set(k, v)
		}
	}

	w := httptest.NewRecorder()
	a.engine.ServeHTTP(w, req)
	return w
}

// mustLogin authenticates and returns the JWT access token for use as a bearer.
func (a *testApp) mustLogin(t *testing.T, email, password string) string {
	t.Helper()
	w := a.request(t, http.MethodPost, "/api/auth/login", map[string]any{
		"login":    email,
		"password": password,
	})
	require.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSON[map[string]string](t, w)
	token := resp["access_token"]
	require.NotEmpty(t, token)
	return token
}

// bearer returns a headers map with the Authorization header set.
func bearer(token string) map[string]string {
	return map[string]string{"Authorization": "Bearer " + token}
}

func decodeJSON[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	require.NoError(t, json.NewDecoder(w.Body).Decode(&v))
	return v
}
