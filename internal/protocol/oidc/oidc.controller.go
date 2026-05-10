package oidc

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/config"
	localcrypto "github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	oidc_dto "github.com/ricardoalcantara/min-idp/internal/protocol/oidc/dto"
)

type OIDCController struct {
	ks     keystore.KeyStore
	issuer string
}

func NewOIDCController(ks keystore.KeyStore, cfg *config.Config) *OIDCController {
	return &OIDCController{ks: ks, issuer: cfg.ExternalURL}
}

func (c *OIDCController) discovery(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, oidc_dto.Build(c.issuer))
}

func (c *OIDCController) jwks(ctx *gin.Context) {
	keys, err := c.ks.PublicKeys(ctx.Request.Context(), keystore_entities.ProtocolOIDC)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	jwks := localcrypto.JWKS{Keys: make([]localcrypto.JWK, 0, len(keys))}
	for _, k := range keys {
		pub, err := localcrypto.ParsePublicKeyPEM([]byte(k.PublicKey))
		if err != nil {
			continue
		}
		switch pk := pub.(type) {
		case *rsa.PublicKey:
			jwks.Keys = append(jwks.Keys, localcrypto.RSAPublicKeyToJWK(pk, k.KID, k.Algorithm))
		case *ecdsa.PublicKey:
			jwks.Keys = append(jwks.Keys, localcrypto.ECPublicKeyToJWK(pk, k.KID, k.Algorithm))
		}
	}
	ctx.JSON(http.StatusOK, jwks)
}

func (c *OIDCController) authorize(ctx *gin.Context)  { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
func (c *OIDCController) token(ctx *gin.Context)       { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
func (c *OIDCController) userinfo(ctx *gin.Context)    { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
func (c *OIDCController) revoke(ctx *gin.Context)      { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
func (c *OIDCController) introspect(ctx *gin.Context)  { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
func (c *OIDCController) logout(ctx *gin.Context)      { ctx.JSON(http.StatusNotImplemented, web.NewMessageDto("not implemented")) }
