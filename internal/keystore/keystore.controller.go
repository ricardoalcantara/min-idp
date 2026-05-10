package keystore

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
)

type KeyStoreController struct {
	ks KeyStore
}

func NewKeyStoreController(ks KeyStore) *KeyStoreController {
	return &KeyStoreController{ks: ks}
}

func (c *KeyStoreController) list(ctx *gin.Context)   { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
func (c *KeyStoreController) rotate(ctx *gin.Context) { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
