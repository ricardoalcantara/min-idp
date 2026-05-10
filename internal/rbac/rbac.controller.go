package rbac

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
)

type RBACController struct {
	service *RBACService
}

func NewRBACController(service *RBACService) *RBACController {
	return &RBACController{service: service}
}

func notImpl(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented"))
}

func (c *RBACController) list(ctx *gin.Context)            { notImpl(ctx) }
func (c *RBACController) create(ctx *gin.Context)           { notImpl(ctx) }
func (c *RBACController) get(ctx *gin.Context)              { notImpl(ctx) }
func (c *RBACController) delete(ctx *gin.Context)           { notImpl(ctx) }
func (c *RBACController) update(ctx *gin.Context)           { notImpl(ctx) }
func (c *RBACController) listPermissions(ctx *gin.Context)  { notImpl(ctx) }
func (c *RBACController) assignPermission(ctx *gin.Context) { notImpl(ctx) }
func (c *RBACController) removePermission(ctx *gin.Context) { notImpl(ctx) }
