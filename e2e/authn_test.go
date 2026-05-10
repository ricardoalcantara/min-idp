package e2e_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthn_Login_Success(t *testing.T) {
	app := setupApp(t)

	w := app.request(t, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    testAdminEmail,
		"password": testAdminPass,
	})

	assert.Equal(t, http.StatusOK, w.Code)

	type loginResp struct {
		SessionID string `json:"session_id"`
	}
	resp := decodeJSON[loginResp](t, w)
	assert.NotEmpty(t, resp.SessionID)
}

func TestAuthn_Login_WrongPassword(t *testing.T) {
	app := setupApp(t)

	w := app.request(t, http.MethodPost, "/api/auth/login", map[string]any{
		"email":    testAdminEmail,
		"password": "wrong-password",
	})

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthn_Login_MissingFields(t *testing.T) {
	app := setupApp(t)

	w := app.request(t, http.MethodPost, "/api/auth/login", map[string]any{
		"email": testAdminEmail,
	})

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthn_Me_Authenticated(t *testing.T) {
	app := setupApp(t)
	cookie := app.mustLogin(t, testAdminEmail, testAdminPass)

	w := app.request(t, http.MethodGet, "/api/me", nil, cookie)

	assert.Equal(t, http.StatusOK, w.Code)
	resp := decodeJSON[map[string]string](t, w)
	assert.Equal(t, testAdminEmail, resp["email"])
}

func TestAuthn_Me_Unauthenticated(t *testing.T) {
	app := setupApp(t)

	w := app.request(t, http.MethodGet, "/api/me", nil)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthn_Logout(t *testing.T) {
	app := setupApp(t)
	cookie := app.mustLogin(t, testAdminEmail, testAdminPass)

	w := app.request(t, http.MethodPost, "/api/auth/logout", nil, cookie)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// session no longer valid
	w = app.request(t, http.MethodGet, "/api/me", nil, cookie)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
