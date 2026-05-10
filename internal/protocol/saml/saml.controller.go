package saml

import (
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
)

type SAMLController struct {
	ks  keystore.KeyStore
	cfg *config.Config
}

func NewSAMLController(ks keystore.KeyStore, cfg *config.Config) *SAMLController {
	return &SAMLController{ks: ks, cfg: cfg}
}

func (c *SAMLController) metadata(ctx *gin.Context) {
	keys, err := c.ks.PublicKeys(ctx.Request.Context(), keystore_entities.ProtocolSAML)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	if len(keys) == 0 {
		ctx.JSON(http.StatusServiceUnavailable, web.NewErrorDto(nil))
		return
	}

	cert := keys[0].Certificate
	cert = strings.ReplaceAll(cert, "-----BEGIN CERTIFICATE-----", "")
	cert = strings.ReplaceAll(cert, "-----END CERTIFICATE-----", "")
	cert = strings.ReplaceAll(cert, "\n", "")
	cert = strings.TrimSpace(cert)

	issuer := c.cfg.ExternalURL
	desc := entityDescriptor{
		EntityID: issuer,
		IDPSSODescriptor: idpSSODescriptor{
			WantAuthnRequestsSigned:    false,
			ProtocolSupportEnumeration: "urn:oasis:names:tc:SAML:2.0:protocol",
			KeyDescriptors: []keyDescriptor{{
				Use: "signing",
				KeyInfo: keyInfo{X509Data: x509Data{X509Certificate: cert}},
			}},
			SingleSignOnServices: []singleSignOnService{
				{Binding: "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST", Location: issuer + "/saml/sso"},
				{Binding: "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect", Location: issuer + "/saml/sso"},
			},
			SingleLogoutServices: []singleLogoutService{
				{Binding: "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST", Location: issuer + "/saml/slo"},
			},
		},
	}

	out, err := xml.MarshalIndent(desc, "", "  ")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	ctx.Data(http.StatusOK, "application/samlmetadata+xml", append([]byte(xml.Header), out...))
}

func (c *SAMLController) sso(ctx *gin.Context) { ctx.JSON(http.StatusNotImplemented, web.NewErrorDto(nil)) }
func (c *SAMLController) slo(ctx *gin.Context) { ctx.JSON(http.StatusNotImplemented, web.NewErrorDto(nil)) }

// XML types
type entityDescriptor struct {
	XMLName          xml.Name         `xml:"urn:oasis:names:tc:SAML:2.0:metadata EntityDescriptor"`
	EntityID         string           `xml:"entityID,attr"`
	IDPSSODescriptor idpSSODescriptor
}
type idpSSODescriptor struct {
	XMLName                    xml.Name             `xml:"urn:oasis:names:tc:SAML:2.0:metadata IDPSSODescriptor"`
	WantAuthnRequestsSigned    bool                 `xml:"WantAuthnRequestsSigned,attr"`
	ProtocolSupportEnumeration string               `xml:"protocolSupportEnumeration,attr"`
	KeyDescriptors             []keyDescriptor
	SingleSignOnServices       []singleSignOnService
	SingleLogoutServices       []singleLogoutService
}
type keyDescriptor struct {
	XMLName xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:metadata KeyDescriptor"`
	Use     string   `xml:"use,attr"`
	KeyInfo keyInfo
}
type keyInfo struct {
	XMLName  xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# KeyInfo"`
	X509Data x509Data
}
type x509Data struct {
	XMLName         xml.Name `xml:"http://www.w3.org/2000/09/xmldsig# X509Data"`
	X509Certificate string   `xml:"http://www.w3.org/2000/09/xmldsig# X509Certificate"`
}
type singleSignOnService struct {
	XMLName  xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:metadata SingleSignOnService"`
	Binding  string   `xml:"Binding,attr"`
	Location string   `xml:"Location,attr"`
}
type singleLogoutService struct {
	XMLName  xml.Name `xml:"urn:oasis:names:tc:SAML:2.0:metadata SingleLogoutService"`
	Binding  string   `xml:"Binding,attr"`
	Location string   `xml:"Location,attr"`
}
