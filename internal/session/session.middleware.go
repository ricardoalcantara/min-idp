package session

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	session_entities "github.com/ricardoalcantara/min-idp/internal/session/entities"
)

type contextKey struct{}

func (s *SessionService) Middleware(cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(cookieName)
		if err == nil && cookie != "" {
			if sess, err := s.Validate(c.Request.Context(), cookie); err == nil {
				c.Set("session", sess)
				c.Request = c.Request.WithContext(
					context.WithValue(c.Request.Context(), contextKey{}, sess),
				)
			}
		}
		c.Next()
	}
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
