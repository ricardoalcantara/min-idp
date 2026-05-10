package rbac

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	rbac_repositories "github.com/ricardoalcantara/min-idp/internal/rbac/repositories"
)

type GroupController struct {
	repo *rbac_repositories.GroupRepository
}

func NewGroupController(repo *rbac_repositories.GroupRepository) *GroupController {
	return &GroupController{repo: repo}
}

func (c *GroupController) list(ctx *gin.Context)   { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
func (c *GroupController) create(ctx *gin.Context)  { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
func (c *GroupController) get(ctx *gin.Context)     { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
func (c *GroupController) update(ctx *gin.Context)  { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
func (c *GroupController) delete(ctx *gin.Context)  { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
