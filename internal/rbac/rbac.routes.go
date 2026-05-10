package rbac

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, c *RBACController, mw ...gin.HandlerFunc) {
	g := r.Group("/api/admin/roles", mw...)
	g.GET("", c.list)
	g.POST("", c.create)
	g.GET("/:id", c.get)
	g.PATCH("/:id", c.update)
	g.DELETE("/:id", c.delete)
	g.GET("/:id/permissions", c.listPermissions)
	g.POST("/:id/permissions", c.assignPermission)
	g.DELETE("/:id/permissions/:permId", c.removePermission)
}
