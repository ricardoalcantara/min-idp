package rbac

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/session"
)

func RequirePermission(rbacSvc *RBACService, perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := session.FromContext(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, web.NewErrorDto(errors.New("authentication required")))
			return
		}
		// Fast path: check JWT claims (no DB hit)
		if claims.HasRole(perm) {
			c.Next()
			return
		}
		// Fallback: DB check (e.g. role assigned after token was issued)
		ok, err := rbacSvc.UserHasPermission(claims.UserID, perm)
		if err != nil || !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, web.NewErrorDto(errors.New("forbidden")))
			return
		}
		c.Next()
	}
}
