package session

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	session_entities "github.com/ricardoalcantara/min-idp/internal/session/entities"
)

type contextKey struct{}

// BearerMiddleware validates the session from the Authorization: Bearer header.
// Used by admin and /api/me routes — pure API, no browser cookies.
func (s *SessionService) BearerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			if sess, err := s.Validate(c.Request.Context(), token); err == nil {
				s.setSession(c, sess)
			}
		}
		c.Next()
	}
}

// CookieMiddleware validates the session from the session cookie.
// Used by browser SSO flows (OIDC authorize, SAML SSO) where the browser
// carries the cookie automatically on redirect.
func (s *SessionService) CookieMiddleware(cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cookie, err := c.Cookie(cookieName); err == nil && cookie != "" {
			if sess, err := s.Validate(c.Request.Context(), cookie); err == nil {
				s.setSession(c, sess)
			}
		}
		c.Next()
	}
}

func (s *SessionService) setSession(c *gin.Context, sess *session_entities.Session) {
	c.Set("session", sess)
	c.Request = c.Request.WithContext(
		context.WithValue(c.Request.Context(), contextKey{}, sess),
	)
}

func RequireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, exists := c.Get("session"); !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, web.NewErrorDto(errors.New("authentication required")))
			return
		}
		c.Next()
	}
}

func FromContext(c *gin.Context) *session_entities.Session {
	raw, _ := c.Get("session")
	sess, _ := raw.(*session_entities.Session)
	return sess
}
