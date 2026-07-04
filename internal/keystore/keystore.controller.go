package keystore

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/go-minstack/web"
	keystore_dto "github.com/ricardoalcantara/min-idp/internal/keystore/dto"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
)

type KeyStoreController struct {
	ks KeyStore
}

func NewKeyStoreController(ks KeyStore) *KeyStoreController {
	return &KeyStoreController{ks: ks}
}

// listAll returns all keys across all protocols from the database.
func (c *KeyStoreController) listAll(ctx *gin.Context) {
	keys, err := c.ks.ListAllKeys(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	dtos := make([]keystore_dto.KeyDto, len(keys))
	for i, k := range keys {
		dtos[i] = keystore_dto.NewKeyDto(k)
	}
	ctx.JSON(http.StatusOK, dtos)
}

// list returns all keys for a specific protocol from the database.
func (c *KeyStoreController) list(ctx *gin.Context) {
	protocol := ctx.Param("protocol")
	if err := validateProtocol(protocol); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	keys, err := c.ks.ListKeysByProtocol(ctx.Request.Context(), protocol)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	dtos := make([]keystore_dto.KeyDto, len(keys))
	for i, k := range keys {
		dtos[i] = keystore_dto.NewKeyDto(k)
	}
	ctx.JSON(http.StatusOK, dtos)
}

func (c *KeyStoreController) rotate(ctx *gin.Context) {
	protocol := ctx.Param("protocol")
	if err := validateProtocol(protocol); err != nil {
		ctx.JSON(http.StatusBadRequest, web.NewErrorDto(err))
		return
	}
	if err := c.ks.GenerateAndRotate(ctx.Request.Context(), protocol); err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	keys, err := c.ks.ListKeysByProtocol(ctx.Request.Context(), protocol)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	dtos := make([]keystore_dto.KeyDto, len(keys))
	for i, k := range keys {
		dtos[i] = keystore_dto.NewKeyDto(k)
	}
	ctx.JSON(http.StatusOK, dtos)
}

func validateProtocol(p string) error {
	if p != keystore_entities.ProtocolOIDC && p != keystore_entities.ProtocolSAML {
		return errors.New("protocol must be 'oidc' or 'saml'")
	}
	return nil
}
