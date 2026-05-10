package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
)

type UserController struct {
	service *UserService
}

func NewUserController(service *UserService) *UserController {
	return &UserController{service: service}
}

func notImpl(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, web.NewErrorDto(nil))
}

func (c *UserController) list(ctx *gin.Context)         { notImpl(ctx) }
func (c *UserController) create(ctx *gin.Context)        { notImpl(ctx) }
func (c *UserController) get(ctx *gin.Context)           { notImpl(ctx) }
func (c *UserController) update(ctx *gin.Context)        { notImpl(ctx) }
func (c *UserController) delete(ctx *gin.Context)        { notImpl(ctx) }
func (c *UserController) assignRole(ctx *gin.Context)    { notImpl(ctx) }
func (c *UserController) removeRole(ctx *gin.Context)    { notImpl(ctx) }
func (c *UserController) assignGroup(ctx *gin.Context)   { notImpl(ctx) }
func (c *UserController) removeGroup(ctx *gin.Context)   { notImpl(ctx) }
func (c *UserController) sessions(ctx *gin.Context)      { notImpl(ctx) }
func (c *UserController) resetPassword(ctx *gin.Context) { notImpl(ctx) }
func (c *UserController) listRoles(ctx *gin.Context)     { notImpl(ctx) }
func (c *UserController) listGroups(ctx *gin.Context)    { notImpl(ctx) }
