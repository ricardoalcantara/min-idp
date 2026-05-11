package sp

import (
	"errors"

	"github.com/go-minstack/repository"
	"github.com/ricardoalcantara/min-idp/internal/types"
	dbpkg "github.com/ricardoalcantara/min-idp/internal/db"
	"github.com/ricardoalcantara/min-idp/internal/crypto"
	sp_dto "github.com/ricardoalcantara/min-idp/internal/sp/dto"
	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
)

var errProtocolMismatch = errors.New("operation not allowed for this SP protocol")

type SPRepository interface {
	Create(sp *sp_entities.ServiceProvider) error
	FindByUUID(uuid string) (*sp_entities.ServiceProvider, error)
	FindBySlug(slug string) (*sp_entities.ServiceProvider, error)
	FindAll(opts ...repository.QueryOption) ([]sp_entities.ServiceProvider, error)
	Update(sp *sp_entities.ServiceProvider) error
	Delete(id uint) error
	GetOIDCClient(spID uint) (*sp_entities.OIDCClient, error)
	UpsertOIDCClient(c *sp_entities.OIDCClient) error
	GetSAMLClient(spID uint) (*sp_entities.SAMLClient, error)
	UpsertSAMLClient(c *sp_entities.SAMLClient) error
	ListAccessRules(spID uint) ([]sp_repositories.AccessRuleRow, error)
	FindSubjectID(subjectType string, entityID uint) (uint, error)
	CreateAccessRule(rule *sp_entities.AccessRule) error
	FindAccessRuleByUUID(uuid string) (*sp_entities.AccessRule, error)
	DeleteAccessRule(id uint) error
}

type SPService struct {
	repo SPRepository
}

func NewSPService(repo SPRepository) *SPService {
	return &SPService{repo: repo}
}

func (s *SPService) Create(slug, name, protocol string) (*sp_entities.ServiceProvider, error) {
	sp := &sp_entities.ServiceProvider{
		Slug:     slug,
		Name:     name,
		Protocol: types.SPProtocol(protocol),
		Enabled:  true,
	}
	return sp, s.repo.Create(sp)
}

func (s *SPService) FindByUUID(id string) (*sp_entities.ServiceProvider, error) {
	return s.repo.FindByUUID(id)
}

func (s *SPService) FindBySlug(slug string) (*sp_entities.ServiceProvider, error) {
	return s.repo.FindBySlug(slug)
}

func (s *SPService) List() ([]sp_entities.ServiceProvider, error) {
	return s.repo.FindAll()
}

func (s *SPService) Update(sp *sp_entities.ServiceProvider, input sp_dto.UpdateSPDto) (*sp_entities.ServiceProvider, error) {
	if input.Name != nil {
		sp.Name = *input.Name
	}
	if input.Enabled != nil {
		sp.Enabled = *input.Enabled
	}
	return sp, s.repo.Update(sp)
}

func (s *SPService) Delete(uuid string) error {
	sp, err := s.repo.FindByUUID(uuid)
	if err != nil {
		return err
	}
	return s.repo.Delete(sp.ID)
}

func (s *SPService) GetOIDCClient(sp *sp_entities.ServiceProvider) (*sp_entities.OIDCClient, error) {
	if sp.Protocol != types.SPProtocolOIDC {
		return nil, errProtocolMismatch
	}
	return s.repo.GetOIDCClient(sp.ID)
}

func (s *SPService) UpsertOIDCClient(sp *sp_entities.ServiceProvider, input sp_dto.UpsertOIDCClientDto) (*sp_entities.OIDCClient, error) {
	if sp.Protocol != types.SPProtocolOIDC {
		return nil, errProtocolMismatch
	}
	client := &sp_entities.OIDCClient{
		SPID:              sp.ID,
		ClientID:          input.ClientID,
		RedirectURIs:      sp_repositories.MarshalStringSlice(input.RedirectURIs),
		GrantTypes:        sp_repositories.MarshalStringSlice(defaultStrings(input.GrantTypes, []string{"authorization_code"})),
		ResponseTypes:     sp_repositories.MarshalStringSlice(defaultStrings(input.ResponseTypes, []string{"code"})),
		Scopes:            sp_repositories.MarshalStringSlice(defaultStrings(input.Scopes, []string{"openid"})),
		TokenEndpointAuth: defaultString(input.TokenEndpointAuth, "client_secret_basic"),
		PKCERequired:      input.PKCERequired,
	}
	if input.ClientSecret != "" {
		from_crypto, err := hashSecret(input.ClientSecret)
		if err != nil {
			return nil, err
		}
		client.ClientSecretHash = from_crypto
	} else {
		existing, err := s.repo.GetOIDCClient(sp.ID)
		if err == nil {
			client.ClientSecretHash = existing.ClientSecretHash
		}
	}
	return client, s.repo.UpsertOIDCClient(client)
}

func (s *SPService) GetSAMLClient(sp *sp_entities.ServiceProvider) (*sp_entities.SAMLClient, error) {
	if sp.Protocol != types.SPProtocolSAML {
		return nil, errProtocolMismatch
	}
	return s.repo.GetSAMLClient(sp.ID)
}

func (s *SPService) UpsertSAMLClient(sp *sp_entities.ServiceProvider, input sp_dto.UpsertSAMLClientDto) (*sp_entities.SAMLClient, error) {
	if sp.Protocol != types.SPProtocolSAML {
		return nil, errProtocolMismatch
	}
	client := &sp_entities.SAMLClient{
		SPID:                 sp.ID,
		EntityID:             input.EntityID,
		ACSURLs:              sp_repositories.MarshalStringSlice(input.ACSURLs),
		SLOUrl:               input.SLOUrl,
		NameIDFormat:         defaultString(input.NameIDFormat, "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"),
		SPCertificate:        input.SPCertificate,
		WantSignedRequests:   input.WantSignedRequests,
		WantSignedAssertions: input.WantSignedAssertions,
	}
	return client, s.repo.UpsertSAMLClient(client)
}

func (s *SPService) ListAccessRules(spID uint) ([]sp_repositories.AccessRuleRow, error) {
	return s.repo.ListAccessRules(spID)
}

func (s *SPService) CreateAccessRule(sp *sp_entities.ServiceProvider, subjectType string, subjectEntityID uint, ruleType string, priority int) (*sp_entities.AccessRule, error) {
	subjectID, err := s.repo.FindSubjectID(subjectType, subjectEntityID)
	if err != nil {
		return nil, err
	}
	rule := &sp_entities.AccessRule{
		SPID:      sp.ID,
		RuleType:  ruleType,
		SubjectID: subjectID,
		Priority:  priority,
	}
	return rule, s.repo.CreateAccessRule(rule)
}

func (s *SPService) DeleteAccessRule(sp *sp_entities.ServiceProvider, ruleUUID string) error {
	rule, err := s.repo.FindAccessRuleByUUID(ruleUUID)
	if err != nil {
		return err
	}
	if rule.SPID != sp.ID {
		return dbpkg.ErrEntityNotFound
	}
	return s.repo.DeleteAccessRule(rule.ID)
}

func hashSecret(secret string) (string, error) {
	return crypto.HashPassword(secret)
}

func defaultString(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func defaultStrings(s, def []string) []string {
	if len(s) == 0 {
		return def
	}
	return s
}
