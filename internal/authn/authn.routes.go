package authn

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, c *AuthnController) {
	r.GET("/login", c.loginPage)

	g := r.Group("/api/auth")
	g.POST("/login", c.login)
	g.POST("/logout", c.logout)
	g.POST("/register", c.register)
}
