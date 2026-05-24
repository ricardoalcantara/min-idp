package saml

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-minstack/web"
	crewjam "github.com/crewjam/saml"
	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	"github.com/ricardoalcantara/min-idp/internal/views"
)

type SAMLController struct {
	samlSvc *SAMLService
	ks      keystore.KeyStore
	cfg     *config.Config
}

func NewSAMLController(samlSvc *SAMLService, ks keystore.KeyStore, cfg *config.Config) *SAMLController {
	return &SAMLController{samlSvc: samlSvc, ks: ks, cfg: cfg}
}

func (c *SAMLController) metadata(ctx *gin.Context) {
	keys, err := c.ks.PublicKeys(ctx.Request.Context(), keystore_entities.ProtocolSAML)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	if len(keys) == 0 {
		ctx.JSON(http.StatusServiceUnavailable, web.NewErrorDto(errors.New("no signing key available")))
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
				Use:     "signing",
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

func (c *SAMLController) sso(ctx *gin.Context) {
	r := ctx.Request
	if r.Method == "POST" {
		r = normalisePostBinding(r)
	}
	idp, err := c.buildIDP(r.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, web.NewErrorDto(err))
		return
	}
	idp.ServeSSO(ctx.Writer, r)
}

// normalisePostBinding detects a DEFLATE-compressed SAMLRequest sent via HTTP-POST
// (non-standard but done by apps like Nextcloud's user_saml). When detected it returns
// a synthetic GET request so crewjam processes it through the Redirect binding path.
func normalisePostBinding(r *http.Request) *http.Request {
	encoded := r.FormValue("SAMLRequest")
	if encoded == "" {
		return r
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return r
	}
	if _, err := io.ReadAll(flate.NewReader(bytes.NewReader(decoded))); err != nil {
		return r // not DEFLATE — standard POST binding, leave unchanged
	}
	q := url.Values{}
	q.Set("SAMLRequest", encoded)
	if rs := r.FormValue("RelayState"); rs != "" {
		q.Set("RelayState", rs)
	}
	clone := r.Clone(r.Context())
	clone.Method = "GET"
	clone.URL.RawQuery = q.Encode()
	return clone
}

func (c *SAMLController) slo(ctx *gin.Context) {
	// Full SP-initiated SLO (SAMLRequest) is out of scope for v1.
	// Handle the simple redirect case: SP clears its session and sends the user
	// here with RelayState (return URL) and entity_id so we can show the logout page.
	relayState := ctx.Query("RelayState")
	entityID := ctx.Query("entity_id")

	spName := c.samlSvc.SPNameByEntityID(entityID)

	views.LogoutTmpl.Render(ctx, map[string]any{
		"SPName":    spName,
		"ReturnURL": relayState,
	})
}

func (c *SAMLController) buildIDP(ctx context.Context) (*crewjam.IdentityProvider, error) {
	key, _, err := c.ks.ActivePrivateKey(ctx, keystore_entities.ProtocolSAML)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("saml: signing key is not RSA")
	}

	keys, err := c.ks.PublicKeys(ctx, keystore_entities.ProtocolSAML)
	if err != nil || len(keys) == 0 {
		return nil, errors.New("saml: no public key available")
	}
	certPEM := keys[0].Certificate
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, errors.New("saml: invalid certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	metaURL, _ := url.Parse(c.cfg.ExternalURL + "/saml/metadata")
	ssoURL, _ := url.Parse(c.cfg.ExternalURL + "/saml/sso")
	sloURL, _ := url.Parse(c.cfg.ExternalURL + "/saml/slo")

	return &crewjam.IdentityProvider{
		Key:                     rsaKey,
		Certificate:             cert,
		MetadataURL:             *metaURL,
		SSOURL:                  *ssoURL,
		LogoutURL:               *sloURL,
		ServiceProviderProvider: c.samlSvc,
		SessionProvider:         c.samlSvc,
		Logger:                  log.Default(),
	}, nil
}

// XML types for metadata endpoint

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
