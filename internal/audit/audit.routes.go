package audit

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, c *AuditController, mw ...gin.HandlerFunc) {
	g := r.Group("/api/admin/audit", mw...)
	g.GET("", c.list)
}
