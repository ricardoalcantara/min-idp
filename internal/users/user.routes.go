package users

import "github.com/gin-gonic/gin"

func RegisterUserRoutes(r *gin.Engine, c *UserController, mw ...gin.HandlerFunc) {
	g := r.Group("/api/admin/users", mw...)
	g.GET("", c.list)
	g.POST("", c.create)
	g.GET("/:id", c.get)
	g.PATCH("/:id", c.update)
	g.DELETE("/:id", c.delete)
	g.POST("/:id/roles", c.assignRole)
	g.DELETE("/:id/roles/:roleId", c.removeRole)
	g.GET("/:id/roles", c.listRoles)
	g.GET("/:id/sessions", c.sessions)
	g.POST("/:id/password", c.resetPassword)
}
