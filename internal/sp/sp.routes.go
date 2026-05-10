package sp

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, c *SPController, mw ...gin.HandlerFunc) {
	g := r.Group("/api/admin/sps", mw...)
	g.GET("", c.list)
	g.POST("", c.create)
	g.GET("/:id", c.get)
	g.PATCH("/:id", c.update)
	g.DELETE("/:id", c.delete)
	g.GET("/:id/oidc", c.getOIDC)
	g.PUT("/:id/oidc", c.putOIDC)
	g.GET("/:id/saml", c.getSAML)
	g.PUT("/:id/saml", c.putSAML)
	g.GET("/:id/access-rules", c.listRules)
	g.POST("/:id/access-rules", c.createRule)
	g.DELETE("/:id/access-rules/:ruleId", c.deleteRule)
}
