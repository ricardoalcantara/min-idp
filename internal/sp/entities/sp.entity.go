package sp_entities

import (
	"github.com/ricardoalcantara/min-idp/internal/db"
	"gorm.io/gorm"
)

type ServiceProvider struct {
	db.Model
	Slug     string `gorm:"uniqueIndex;not null"`
	Name     string `gorm:"not null"`
	Protocol string `gorm:"not null"`
	Enabled  bool   `gorm:"default:true"`
}

type OIDCClient struct {
	gorm.Model
	SPID              uint   `gorm:"uniqueIndex;not null"`
	ClientID          string `gorm:"uniqueIndex;not null"`
	ClientSecretHash  string `gorm:"not null"`
	RedirectURIs      string `gorm:"default:'[]'"`
	GrantTypes        string `gorm:"default:'[\"authorization_code\"]'"`
	ResponseTypes     string `gorm:"default:'[\"code\"]'"`
	Scopes            string `gorm:"default:'[\"openid\"]'"`
	TokenEndpointAuth string `gorm:"default:'client_secret_basic'"`
	PKCERequired      bool   `gorm:"default:false"`
}

type SAMLClient struct {
	gorm.Model
	SPID                 uint   `gorm:"uniqueIndex;not null"`
	EntityID             string `gorm:"uniqueIndex;not null"`
	ACSURLs              string `gorm:"default:'[]'"`
	SLOUrl               string
	NameIDFormat         string `gorm:"default:'urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress'"`
	SPCertificate        string
	WantSignedRequests   bool `gorm:"default:false"`
	WantSignedAssertions bool `gorm:"default:true"`
}

type AccessRule struct {
	gorm.Model
	SPID        uint   `gorm:"not null;index"`
	RuleType    string `gorm:"not null"`
	SubjectType string `gorm:"not null"`
	SubjectID   uint   `gorm:"not null"`
	Priority    int    `gorm:"default:0"`
}
