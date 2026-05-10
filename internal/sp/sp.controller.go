package sp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
)

type SPController struct {
	service *SPService
}

func NewSPController(service *SPService) *SPController {
	return &SPController{service: service}
}

func notImpl(ctx *gin.Context) {
	ctx.JSON(http.StatusNotImplemented, web.NewErrorDto(nil))
}

func (c *SPController) list(ctx *gin.Context)      { notImpl(ctx) }
func (c *SPController) create(ctx *gin.Context)     { notImpl(ctx) }
func (c *SPController) get(ctx *gin.Context)        { notImpl(ctx) }
func (c *SPController) update(ctx *gin.Context)     { notImpl(ctx) }
func (c *SPController) delete(ctx *gin.Context)     { notImpl(ctx) }
func (c *SPController) getOIDC(ctx *gin.Context)    { notImpl(ctx) }
func (c *SPController) putOIDC(ctx *gin.Context)    { notImpl(ctx) }
func (c *SPController) getSAML(ctx *gin.Context)    { notImpl(ctx) }
func (c *SPController) putSAML(ctx *gin.Context)    { notImpl(ctx) }
func (c *SPController) listRules(ctx *gin.Context)  { notImpl(ctx) }
func (c *SPController) createRule(ctx *gin.Context) { notImpl(ctx) }
func (c *SPController) deleteRule(ctx *gin.Context) { notImpl(ctx) }
