package oidc

import (
	"testing"

	"github.com/go-minstack/go-minstack/repository"
	localcrypto "github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/db"
	"github.com/ricardoalcantara/min-idp/internal/sp"
	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockOIDCSPRepo struct {
	client *sp_entities.OIDCClient
	err    error
}

func (m *mockOIDCSPRepo) Create(_ *sp_entities.ServiceProvider) error { return m.err }
func (m *mockOIDCSPRepo) FindByUUID(_ string) (*sp_entities.ServiceProvider, error) {
	return nil, m.err
}
func (m *mockOIDCSPRepo) FindBySlug(_ string) (*sp_entities.ServiceProvider, error) {
	return nil, m.err
}
func (m *mockOIDCSPRepo) FindByID(_ uint) (*sp_entities.ServiceProvider, error) { return nil, m.err }
func (m *mockOIDCSPRepo) FindAll(_ ...repository.QueryOption) ([]sp_entities.ServiceProvider, error) {
	return nil, m.err
}
func (m *mockOIDCSPRepo) Update(_ *sp_entities.ServiceProvider) error { return m.err }
func (m *mockOIDCSPRepo) Delete(_ uint) error                         { return m.err }
func (m *mockOIDCSPRepo) GetOIDCClient(_ uint) (*sp_entities.OIDCClient, error) {
	return nil, m.err
}
func (m *mockOIDCSPRepo) FindOIDCClientByClientID(_ string) (*sp_entities.OIDCClient, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.client, nil
}
func (m *mockOIDCSPRepo) UpsertOIDCClient(_ *sp_entities.OIDCClient) error { return m.err }
func (m *mockOIDCSPRepo) GetSAMLClient(_ uint) (*sp_entities.SAMLClient, error) {
	return nil, m.err
}
func (m *mockOIDCSPRepo) FindSAMLClientByEntityID(_ string) (*sp_entities.SAMLClient, error) {
	return nil, m.err
}
func (m *mockOIDCSPRepo) UpsertSAMLClient(_ *sp_entities.SAMLClient) error { return m.err }
func (m *mockOIDCSPRepo) ListAccessRules(_ uint) ([]sp_repositories.AccessRuleRow, error) {
	return nil, m.err
}
func (m *mockOIDCSPRepo) FindSubjectID(_ string, _ uint) (uint, error)     { return 0, m.err }
func (m *mockOIDCSPRepo) CreateAccessRule(_ *sp_entities.AccessRule) error { return m.err }
func (m *mockOIDCSPRepo) FindAccessRuleByUUID(_ string) (*sp_entities.AccessRule, error) {
	return nil, m.err
}
func (m *mockOIDCSPRepo) DeleteAccessRule(_ uint) error { return m.err }

var _ sp.SPRepository = (*mockOIDCSPRepo)(nil)

func newTestOIDCService(repo *mockOIDCSPRepo) *OIDCService {
	return &OIDCService{spRepo: repo}
}

func TestOIDCService_ValidateClient_Public_NoSecret(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{
		client: &sp_entities.OIDCClient{
			ClientID:          "public-spa",
			TokenEndpointAuth: "none",
			RedirectURIs:      `["http://localhost:5173/callback"]`,
		},
	})

	client, err := svc.ValidateClient("public-spa", "", "http://localhost:5173/callback")
	require.NoError(t, err)
	assert.Equal(t, "none", client.TokenEndpointAuth)
}

func TestOIDCService_ValidateClient_Public_RejectsSecret(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{
		client: &sp_entities.OIDCClient{
			ClientID:          "public-spa",
			TokenEndpointAuth: "none",
		},
	})

	_, err := svc.ValidateClient("public-spa", "unexpected-secret", "")
	assert.ErrorIs(t, err, ErrInvalidClient)
}

func TestOIDCService_ValidateClient_Confidential_RequiresSecret(t *testing.T) {
	hash, err := localcrypto.HashPassword("super-secret")
	require.NoError(t, err)

	svc := newTestOIDCService(&mockOIDCSPRepo{
		client: &sp_entities.OIDCClient{
			ClientID:          "confidential-app",
			TokenEndpointAuth: "client_secret_basic",
			ClientSecretHash:  hash,
		},
	})

	_, err = svc.ValidateClient("confidential-app", "", "")
	assert.ErrorIs(t, err, ErrInvalidClient)
}

func TestOIDCService_ValidateClient_Confidential_WrongSecret(t *testing.T) {
	hash, err := localcrypto.HashPassword("super-secret")
	require.NoError(t, err)

	svc := newTestOIDCService(&mockOIDCSPRepo{
		client: &sp_entities.OIDCClient{
			ClientID:          "confidential-app",
			TokenEndpointAuth: "client_secret_basic",
			ClientSecretHash:  hash,
		},
	})

	_, err = svc.ValidateClient("confidential-app", "wrong-secret", "")
	assert.ErrorIs(t, err, ErrInvalidClient)
}

func TestOIDCService_ValidateClient_Confidential_ValidSecret(t *testing.T) {
	hash, err := localcrypto.HashPassword("super-secret")
	require.NoError(t, err)

	svc := newTestOIDCService(&mockOIDCSPRepo{
		client: &sp_entities.OIDCClient{
			ClientID:          "confidential-app",
			TokenEndpointAuth: "client_secret_basic",
			ClientSecretHash:  hash,
		},
	})

	client, err := svc.ValidateClient("confidential-app", "super-secret", "")
	require.NoError(t, err)
	assert.Equal(t, "confidential-app", client.ClientID)
}

func TestOIDCService_ValidateAuthorizeRequest_PKCERequired_MissingChallenge(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{})
	client := &sp_entities.OIDCClient{PKCERequired: true}

	err := svc.ValidateAuthorizeRequest(client, "", "S256")
	assert.ErrorIs(t, err, ErrInvalidRequest)
}

func TestOIDCService_ValidateAuthorizeRequest_PKCERequired_RejectPlain(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{})
	client := &sp_entities.OIDCClient{PKCERequired: true}

	err := svc.ValidateAuthorizeRequest(client, "challenge", "plain")
	assert.ErrorIs(t, err, ErrInvalidRequest)
}

func TestOIDCService_ValidateAuthorizeRequest_PKCERequired_S256(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{})
	client := &sp_entities.OIDCClient{PKCERequired: true}

	err := svc.ValidateAuthorizeRequest(client, "challenge", "S256")
	assert.NoError(t, err)
}

func TestOIDCService_ValidateAuthorizeRequest_PKCEOptional_NoChallenge(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{})
	client := &sp_entities.OIDCClient{PKCERequired: false}

	err := svc.ValidateAuthorizeRequest(client, "", "")
	assert.NoError(t, err)
}

func TestOIDCService_ValidateClient_NotFound(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{err: db.ErrEntityNotFound})
	_, err := svc.ValidateClient("missing", "", "")
	assert.ErrorIs(t, err, ErrInvalidClient)
}

func TestOIDCService_ValidatePostLogoutRedirectURI_Allowed(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{})
	client := &sp_entities.OIDCClient{
		PostLogoutRedirectURIs: `["http://localhost:5173/"]`,
	}
	err := svc.ValidatePostLogoutRedirectURI(client, "http://localhost:5173/")
	assert.NoError(t, err)
}

func TestOIDCService_ValidatePostLogoutRedirectURI_NotInList(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{})
	client := &sp_entities.OIDCClient{
		PostLogoutRedirectURIs: `["http://localhost:5173/"]`,
	}
	err := svc.ValidatePostLogoutRedirectURI(client, "https://evil.example/")
	assert.ErrorIs(t, err, ErrInvalidPostLogoutRedirectURI)
}

func TestOIDCService_ValidatePostLogoutRedirectURI_EmptyURI(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{})
	client := &sp_entities.OIDCClient{PostLogoutRedirectURIs: `[]`}
	err := svc.ValidatePostLogoutRedirectURI(client, "")
	assert.NoError(t, err)
}

func TestOIDCService_ValidatePostLogoutRedirectURI_EmptyAllowlist(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{})
	client := &sp_entities.OIDCClient{PostLogoutRedirectURIs: `[]`}
	err := svc.ValidatePostLogoutRedirectURI(client, "http://localhost:5173/")
	assert.ErrorIs(t, err, ErrInvalidPostLogoutRedirectURI)
}

func TestOIDCService_ValidatePostLogoutRedirectURI_UnsafeScheme(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{})
	client := &sp_entities.OIDCClient{
		PostLogoutRedirectURIs: `["javascript:alert(1)"]`,
	}
	err := svc.ValidatePostLogoutRedirectURI(client, "javascript:alert(1)")
	assert.ErrorIs(t, err, ErrInvalidPostLogoutRedirectURI)
}

func TestOIDCService_ResolveLogoutClient_FromClientID(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{
		client: &sp_entities.OIDCClient{ClientID: "my-app"},
	})
	client, err := svc.ResolveLogoutClient("my-app", "")
	require.NoError(t, err)
	assert.Equal(t, "my-app", client.ClientID)
}
