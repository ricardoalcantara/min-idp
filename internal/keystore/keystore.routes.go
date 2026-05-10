package keystore

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, c *KeyStoreController, mw ...gin.HandlerFunc) {
	g := r.Group("/api/admin/keys", mw...)
	g.GET("/:protocol", c.list)
	g.POST("/:protocol/rotate", c.rotate)
}
