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
	client         *sp_entities.OIDCClient
	err            error
	upsertedClient *sp_entities.OIDCClient
	upsertCalls    int
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
func (m *mockOIDCSPRepo) UpsertOIDCClient(c *sp_entities.OIDCClient) error {
	m.upsertCalls++
	if m.err != nil {
		return m.err
	}
	m.upsertedClient = c
	m.client = c
	return nil
}
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

func TestOIDCService_ValidateOrDiscoverPostLogoutRedirectURI_DiscoverSameOrigin(t *testing.T) {
	repo := &mockOIDCSPRepo{}
	svc := newTestOIDCService(repo)
	client := &sp_entities.OIDCClient{
		RedirectURIs:           `["http://localhost:5173/callback"]`,
		PostLogoutRedirectURIs: `[]`,
	}

	err := svc.ValidateOrDiscoverPostLogoutRedirectURI(client, "http://localhost:5173/")
	require.NoError(t, err)
	assert.Equal(t, 1, repo.upsertCalls)
	require.NotNil(t, repo.upsertedClient)
	assert.Equal(t, `["http://localhost:5173/"]`, repo.upsertedClient.PostLogoutRedirectURIs)
	assert.Equal(t, `["http://localhost:5173/"]`, client.PostLogoutRedirectURIs)
}

func TestOIDCService_ValidateOrDiscoverPostLogoutRedirectURI_DifferentPort(t *testing.T) {
	repo := &mockOIDCSPRepo{}
	svc := newTestOIDCService(repo)
	client := &sp_entities.OIDCClient{
		RedirectURIs:           `["http://localhost:5173/callback"]`,
		PostLogoutRedirectURIs: `[]`,
	}

	err := svc.ValidateOrDiscoverPostLogoutRedirectURI(client, "http://localhost:3000/")
	assert.ErrorIs(t, err, ErrInvalidPostLogoutRedirectURI)
	assert.Equal(t, 0, repo.upsertCalls)
}

func TestOIDCService_ValidateOrDiscoverPostLogoutRedirectURI_DifferentDomain(t *testing.T) {
	repo := &mockOIDCSPRepo{}
	svc := newTestOIDCService(repo)
	client := &sp_entities.OIDCClient{
		RedirectURIs:           `["http://localhost:5173/callback"]`,
		PostLogoutRedirectURIs: `[]`,
	}

	err := svc.ValidateOrDiscoverPostLogoutRedirectURI(client, "https://evil.example/")
	assert.ErrorIs(t, err, ErrInvalidPostLogoutRedirectURI)
	assert.Equal(t, 0, repo.upsertCalls)
}

func TestOIDCService_ValidateOrDiscoverPostLogoutRedirectURI_EmptyRedirectURIs(t *testing.T) {
	repo := &mockOIDCSPRepo{}
	svc := newTestOIDCService(repo)
	client := &sp_entities.OIDCClient{
		RedirectURIs:           `[]`,
		PostLogoutRedirectURIs: `[]`,
	}

	err := svc.ValidateOrDiscoverPostLogoutRedirectURI(client, "http://localhost:5173/")
	assert.ErrorIs(t, err, ErrInvalidPostLogoutRedirectURI)
	assert.Equal(t, 0, repo.upsertCalls)
}

func TestOIDCService_ValidateOrDiscoverPostLogoutRedirectURI_NonEmptyAllowlistNoDiscovery(t *testing.T) {
	repo := &mockOIDCSPRepo{}
	svc := newTestOIDCService(repo)
	client := &sp_entities.OIDCClient{
		RedirectURIs:           `["http://localhost:5173/callback"]`,
		PostLogoutRedirectURIs: `["http://localhost:5173/"]`,
	}

	err := svc.ValidateOrDiscoverPostLogoutRedirectURI(client, "http://localhost:5173/after-logout")
	assert.ErrorIs(t, err, ErrInvalidPostLogoutRedirectURI)
	assert.Equal(t, 0, repo.upsertCalls)
}

func TestOIDCService_ValidateOrDiscoverPostLogoutRedirectURI_DefaultPortNormalization(t *testing.T) {
	repo := &mockOIDCSPRepo{}
	svc := newTestOIDCService(repo)
	client := &sp_entities.OIDCClient{
		RedirectURIs:           `["https://app.example.com/cb"]`,
		PostLogoutRedirectURIs: `[]`,
	}

	err := svc.ValidateOrDiscoverPostLogoutRedirectURI(client, "https://app.example.com:443/done")
	require.NoError(t, err)
	assert.Equal(t, 1, repo.upsertCalls)
	require.NotNil(t, repo.upsertedClient)
	assert.Equal(t, `["https://app.example.com:443/done"]`, repo.upsertedClient.PostLogoutRedirectURIs)
}

func TestOIDCService_ResolveLogoutClient_FromClientID(t *testing.T) {
	svc := newTestOIDCService(&mockOIDCSPRepo{
		client: &sp_entities.OIDCClient{ClientID: "my-app"},
	})
	client, err := svc.ResolveLogoutClient("my-app", "")
	require.NoError(t, err)
	assert.Equal(t, "my-app", client.ClientID)
}
