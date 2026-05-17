package saml

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"net/http"
	"net/url"
	"os"

	crewjam "github.com/crewjam/saml"
	dbpkg "github.com/ricardoalcantara/min-idp/internal/db"
	"github.com/ricardoalcantara/min-idp/internal/session"
	"github.com/ricardoalcantara/min-idp/internal/sp"
	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
)

// SAMLService implements crewjam/saml's ServiceProviderProvider and SessionProvider interfaces.
type SAMLService struct {
	spRepo sp.SPRepository
	gate   *sp.SPGateService
}

func NewSAMLService(spRepo sp.SPRepository, gate *sp.SPGateService) *SAMLService {
	return &SAMLService{spRepo: spRepo, gate: gate}
}

// SPNameByEntityID resolves the display name of a registered SAML SP.
// Returns a fallback string if the entity ID is unknown.
func (s *SAMLService) SPNameByEntityID(entityID string) string {
	client, err := s.spRepo.FindSAMLClientByEntityID(entityID)
	if err != nil {
		return "the application"
	}
	sp, err := s.spRepo.FindByID(client.SPID)
	if err != nil {
		return "the application"
	}
	return sp.Name
}

// GetServiceProvider implements saml.ServiceProviderProvider.
// Looks up the registered SAMLClient by entity ID and builds a crewjam EntityDescriptor.
func (s *SAMLService) GetServiceProvider(r *http.Request, entityID string) (*crewjam.EntityDescriptor, error) {
	client, err := s.spRepo.FindSAMLClientByEntityID(entityID)
	if errors.Is(err, dbpkg.ErrEntityNotFound) {
		return nil, os.ErrNotExist
	}
	if err != nil {
		return nil, err
	}
	return s.buildEntityDescriptor(client), nil
}

// GetSession implements saml.SessionProvider.
// Returns session claims if the user is authenticated and permitted, or redirects to login.
func (s *SAMLService) GetSession(w http.ResponseWriter, r *http.Request, req *crewjam.IdpAuthnRequest) *crewjam.Session {
	claims := session.FromRequest(r)
	if claims == nil {
		http.Redirect(w, r, "/login?next="+url.QueryEscape(r.URL.String()), http.StatusFound)
		return nil
	}

	entityID := req.Request.Issuer.Value
	client, err := s.spRepo.FindSAMLClientByEntityID(entityID)
	if err != nil {
		http.Error(w, "unknown service provider", http.StatusForbidden)
		return nil
	}

	spRecord, err := s.spRepo.FindByID(client.SPID)
	if err != nil {
		http.Error(w, "unknown service provider", http.StatusForbidden)
		return nil
	}

	if err := s.gate.CanSSO(claims.UserID, spRecord); err != nil {
		http.Error(w, "access denied", http.StatusForbidden)
		return nil
	}

	roleValues := make([]crewjam.AttributeValue, len(claims.Roles))
	for i, r := range claims.Roles {
		roleValues[i] = crewjam.AttributeValue{Type: "xs:string", Value: r}
	}

	nameID, nameIDFormat := resolveNameID(claims, client.NameIDFormat)

	return &crewjam.Session{
		ID:           claims.SessionUUID,
		Index:        claims.SessionUUID,
		NameID:       nameID,
		NameIDFormat: nameIDFormat,
		CustomAttributes: []crewjam.Attribute{
			{
				FriendlyName: "sub",
				Name:         "sub",
				NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
				Values:       []crewjam.AttributeValue{{Type: "xs:string", Value: claims.UserUUID}},
			},
			{
				FriendlyName: "email",
				Name:         "email",
				NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
				Values:       []crewjam.AttributeValue{{Type: "xs:string", Value: claims.Email}},
			},
			{
				FriendlyName: "username",
				Name:         "username",
				NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
				Values:       []crewjam.AttributeValue{{Type: "xs:string", Value: func() string { if claims.Username != "" { return claims.Username }; return claims.Email }()}},
			},
			{
				FriendlyName: "name",
				Name:         "name",
				NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
				Values:       []crewjam.AttributeValue{{Type: "xs:string", Value: func() string { if claims.Name != "" { return claims.Name }; return claims.Email }()}},
			},
			{
				FriendlyName: "roles",
				Name:         "roles",
				NameFormat:   "urn:oasis:names:tc:SAML:2.0:attrname-format:basic",
				Values:       roleValues,
			},
		},
	}
}

func (s *SAMLService) buildEntityDescriptor(client *sp_entities.SAMLClient) *crewjam.EntityDescriptor {
	acsURLs := sp_repositories.UnmarshalStringSlice(client.ACSURLs)

	var acs []crewjam.IndexedEndpoint
	for i, u := range acsURLs {
		acs = append(acs, crewjam.IndexedEndpoint{
			Binding:  crewjam.HTTPPostBinding,
			Location: u,
			Index:    i,
		})
	}

	wantSigned := client.WantSignedAssertions
	desc := &crewjam.EntityDescriptor{
		EntityID: client.EntityID,
		SPSSODescriptors: []crewjam.SPSSODescriptor{
			{
				SSODescriptor: crewjam.SSODescriptor{
					RoleDescriptor: crewjam.RoleDescriptor{},
					NameIDFormats:  []crewjam.NameIDFormat{crewjam.NameIDFormat(client.NameIDFormat)},
				},
				WantAssertionsSigned:      &wantSigned,
				AssertionConsumerServices: acs,
			},
		},
	}

	// Add SP certificate if provided (used to verify signed AuthnRequests)
	if client.SPCertificate != "" {
		certBlock, _ := pem.Decode([]byte(client.SPCertificate))
		if certBlock != nil {
			if cert, err := x509.ParseCertificate(certBlock.Bytes); err == nil {
				kd := crewjam.KeyDescriptor{
					Use: "signing",
					KeyInfo: crewjam.KeyInfo{
						X509Data: crewjam.X509Data{
							X509Certificates: []crewjam.X509Certificate{
								{Data: base64.StdEncoding.EncodeToString(cert.Raw)},
							},
						},
					},
				}
				desc.SPSSODescriptors[0].KeyDescriptors = append(desc.SPSSODescriptors[0].KeyDescriptors, kd)
			}
		}
	}

	return desc
}

func resolveNameID(claims *session.SessionClaims, format string) (nameID, nameIDFormat string) {
	switch format {
	case "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress":
		return claims.Email, format
	case "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified":
		return claims.Username, format
	default:
		return claims.UserUUID, "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
	}
}
