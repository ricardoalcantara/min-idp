package rbac

import "github.com/gin-gonic/gin"

func RegisterGroupRoutes(r *gin.Engine, c *GroupController, mw ...gin.HandlerFunc) {
	g := r.Group("/api/admin/groups", mw...)
	g.GET("", c.list)
	g.POST("", c.create)
	g.GET("/:id", c.get)
	g.PATCH("/:id", c.update)
	g.DELETE("/:id", c.delete)
}
