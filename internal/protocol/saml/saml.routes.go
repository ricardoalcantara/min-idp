package saml

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, c *SAMLController) {
	g := r.Group("/saml")
	g.GET("/metadata", c.metadata)
	g.GET("/sso", c.sso)
	g.POST("/sso", c.sso)
	g.GET("/slo", c.slo)
	g.POST("/slo", c.slo)
}
