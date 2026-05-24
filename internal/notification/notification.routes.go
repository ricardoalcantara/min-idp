package notification

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.Engine, c *NotificationController, mw ...gin.HandlerFunc) {
	g := r.Group("/api/admin/notifications", mw...)
	g.POST("/test", c.sendTest)
}
