package rbac

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/session"
)

func RequirePermission(rbacSvc *RBACService, perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sess := session.FromContext(c)
		if sess == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, web.NewErrorDto(nil))
			return
		}
		ok, err := rbacSvc.UserHasPermission(sess.UserID, perm)
		if err != nil || !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, web.NewErrorDto(nil))
			return
		}
		c.Next()
	}
}
