package sp_dto

import (
	"time"

	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
	"github.com/ricardoalcantara/min-idp/internal/types"
)

// --- ServiceProvider ---

type SPDto struct {
	ID        string           `json:"id"`
	Slug      string           `json:"slug"`
	Name      string           `json:"name"`
	Protocol  types.SPProtocol `json:"protocol"`
	Enabled   bool             `json:"enabled"`
	CreatedAt time.Time        `json:"created_at"`
}

func NewSPDto(s *sp_entities.ServiceProvider) SPDto {
	return SPDto{
		ID:        s.UUID.String(),
		Slug:      s.Slug,
		Name:      s.Name,
		Protocol:  s.Protocol,
		Enabled:   s.Enabled,
		CreatedAt: s.CreatedAt,
	}
}

type CreateSPDto struct {
	Slug     string `json:"slug"     binding:"required"`
	Name     string `json:"name"     binding:"required"`
	Protocol string `json:"protocol" binding:"required,oneof=oidc saml"`
}

type UpdateSPDto struct {
	Name    *string `json:"name"`
	Enabled *bool   `json:"enabled"`
}

// --- OIDC Client ---

type OIDCClientDto struct {
	ClientID               string   `json:"client_id"`
	RedirectURIs           []string `json:"redirect_uris"`
	PostLogoutRedirectURIs []string `json:"post_logout_redirect_uris"`
	GrantTypes             []string `json:"grant_types"`
	ResponseTypes          []string `json:"response_types"`
	Scopes                 []string `json:"scopes"`
	TokenEndpointAuth      string   `json:"token_endpoint_auth"`
	PKCERequired           bool     `json:"pkce_required"`
}

type UpsertOIDCClientDto struct {
	ClientID               string   `json:"client_id"                    binding:"required"`
	ClientSecret           string   `json:"client_secret"`
	RedirectURIs           []string `json:"redirect_uris"                binding:"required"`
	PostLogoutRedirectURIs []string `json:"post_logout_redirect_uris"    binding:"required"`
	GrantTypes             []string `json:"grant_types"`
	ResponseTypes          []string `json:"response_types"`
	Scopes                 []string `json:"scopes"`
	TokenEndpointAuth      string   `json:"token_endpoint_auth"`
	PKCERequired           bool     `json:"pkce_required"`
}

// --- SAML Client ---

type SAMLClientDto struct {
	EntityID             string   `json:"entity_id"`
	ACSURLs              []string `json:"acs_urls"`
	SLOUrl               string   `json:"slo_url,omitempty"`
	NameIDFormat         string   `json:"name_id_format"`
	WantSignedRequests   bool     `json:"want_signed_requests"`
	WantSignedAssertions bool     `json:"want_signed_assertions"`
}

type UpsertSAMLClientDto struct {
	EntityID             string   `json:"entity_id"              binding:"required"`
	ACSURLs              []string `json:"acs_urls"               binding:"required"`
	SLOUrl               string   `json:"slo_url"`
	NameIDFormat         string   `json:"name_id_format"`
	SPCertificate        string   `json:"sp_certificate"`
	WantSignedRequests   bool     `json:"want_signed_requests"`
	WantSignedAssertions bool     `json:"want_signed_assertions"`
}

// --- Access Rules ---

type AccessRuleDto struct {
	ID          string `json:"id"`
	RuleType    string `json:"rule_type"`
	SubjectType string `json:"subject_type"`
	SubjectID   string `json:"subject_id"` // UUID of the subject
	Priority    int    `json:"priority"`
}

type CreateAccessRuleDto struct {
	RuleType    string `json:"rule_type"    binding:"required,oneof=allow deny"`
	SubjectType string `json:"subject_type" binding:"required,oneof=group role user"`
	SubjectID   string `json:"subject_id"   binding:"required"` // UUID of the subject
	Priority    int    `json:"priority"`
}
