package sp

import (
	"errors"
	"testing"

	"github.com/go-minstack/go-minstack/repository"
	"github.com/ricardoalcantara/min-idp/internal/db"
	sp_dto "github.com/ricardoalcantara/min-idp/internal/sp/dto"
	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
	"github.com/ricardoalcantara/min-idp/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock ---

type mockSPRepo struct {
	sp  *sp_entities.ServiceProvider
	sps []sp_entities.ServiceProvider
	err error
}

func (m *mockSPRepo) Create(_ *sp_entities.ServiceProvider) error               { return m.err }
func (m *mockSPRepo) Update(_ *sp_entities.ServiceProvider) error               { return m.err }
func (m *mockSPRepo) Delete(_ uint) error                                       { return m.err }
func (m *mockSPRepo) FindByUUID(_ string) (*sp_entities.ServiceProvider, error) { return m.sp, m.err }
func (m *mockSPRepo) FindBySlug(_ string) (*sp_entities.ServiceProvider, error) { return m.sp, m.err }
func (m *mockSPRepo) FindByID(_ uint) (*sp_entities.ServiceProvider, error)     { return m.sp, m.err }
func (m *mockSPRepo) FindAll(_ ...repository.QueryOption) ([]sp_entities.ServiceProvider, error) {
	return m.sps, m.err
}

func (m *mockSPRepo) GetOIDCClient(_ uint) (*sp_entities.OIDCClient, error) { return nil, m.err }
func (m *mockSPRepo) FindOIDCClientByClientID(_ string) (*sp_entities.OIDCClient, error) {
	return nil, m.err
}
func (m *mockSPRepo) UpsertOIDCClient(_ *sp_entities.OIDCClient) error { return m.err }
func (m *mockSPRepo) FindSAMLClientByEntityID(_ string) (*sp_entities.SAMLClient, error) {
	return nil, m.err
}
func (m *mockSPRepo) GetSAMLClient(_ uint) (*sp_entities.SAMLClient, error) { return nil, m.err }
func (m *mockSPRepo) UpsertSAMLClient(_ *sp_entities.SAMLClient) error      { return m.err }
func (m *mockSPRepo) ListAccessRules(_ uint) ([]sp_repositories.AccessRuleRow, error) {
	return nil, m.err
}
func (m *mockSPRepo) FindSubjectID(_ string, _ uint) (uint, error)     { return 0, m.err }
func (m *mockSPRepo) CreateAccessRule(_ *sp_entities.AccessRule) error { return m.err }
func (m *mockSPRepo) FindAccessRuleByUUID(_ string) (*sp_entities.AccessRule, error) {
	return nil, m.err
}
func (m *mockSPRepo) DeleteAccessRule(_ uint) error { return m.err }

func newTestSPSvc(repo SPRepository) *SPService {
	return NewSPService(repo)
}

// --- Create ---

func TestSPService_Create_Success(t *testing.T) {
	svc := newTestSPSvc(&mockSPRepo{})
	sp, err := svc.Create("my-app", "My App", "oidc")
	require.NoError(t, err)
	assert.Equal(t, "my-app", sp.Slug)
	assert.Equal(t, "My App", sp.Name)
	assert.True(t, sp.Enabled)
}

func TestSPService_Create_RepoError(t *testing.T) {
	svc := newTestSPSvc(&mockSPRepo{err: errors.New("db down")})
	_, err := svc.Create("my-app", "My App", "oidc")
	assert.Error(t, err)
}

// --- FindByUUID ---

func TestSPService_FindByUUID_Success(t *testing.T) {
	sp := &sp_entities.ServiceProvider{Slug: "my-app"}
	svc := newTestSPSvc(&mockSPRepo{sp: sp})
	got, err := svc.FindByUUID("some-uuid")
	require.NoError(t, err)
	assert.Equal(t, "my-app", got.Slug)
}

func TestSPService_FindByUUID_NotFound(t *testing.T) {
	svc := newTestSPSvc(&mockSPRepo{err: db.ErrEntityNotFound})
	_, err := svc.FindByUUID("bad-uuid")
	assert.ErrorIs(t, err, db.ErrEntityNotFound)
}

// --- List ---

func TestSPService_List_Empty(t *testing.T) {
	svc := newTestSPSvc(&mockSPRepo{sps: []sp_entities.ServiceProvider{}})
	sps, err := svc.List()
	require.NoError(t, err)
	assert.Empty(t, sps)
}

func TestSPService_List_RepoError(t *testing.T) {
	svc := newTestSPSvc(&mockSPRepo{err: errors.New("db down")})
	_, err := svc.List()
	assert.Error(t, err)
}

// --- UpsertOIDCClient ---

type upsertOIDCRepo struct {
	mockSPRepo
	oidcClient *sp_entities.OIDCClient
	upserted   *sp_entities.OIDCClient
}

func (m *upsertOIDCRepo) GetOIDCClient(_ uint) (*sp_entities.OIDCClient, error) {
	if m.oidcClient == nil {
		return nil, db.ErrEntityNotFound
	}
	return m.oidcClient, nil
}

func (m *upsertOIDCRepo) UpsertOIDCClient(c *sp_entities.OIDCClient) error {
	m.upserted = c
	return nil
}

func oidcSP() *sp_entities.ServiceProvider {
	sp := &sp_entities.ServiceProvider{Protocol: types.SPProtocolOIDC}
	sp.ID = 1
	return sp
}

func TestSPService_UpsertOIDCClient_PublicClient_NoSecret(t *testing.T) {
	repo := &upsertOIDCRepo{}
	svc := newTestSPSvc(repo)

	client, err := svc.UpsertOIDCClient(oidcSP(), sp_dto.UpsertOIDCClientDto{
		ClientID:          "public-spa",
		RedirectURIs:      []string{"http://localhost:5173/callback"},
		TokenEndpointAuth: "none",
	})
	require.NoError(t, err)
	assert.Equal(t, "none", client.TokenEndpointAuth)
	assert.True(t, client.PKCERequired)
	assert.Empty(t, client.ClientSecretHash)
}

func TestSPService_UpsertOIDCClient_Confidential_RequiresSecretOnCreate(t *testing.T) {
	repo := &upsertOIDCRepo{}
	svc := newTestSPSvc(repo)

	_, err := svc.UpsertOIDCClient(oidcSP(), sp_dto.UpsertOIDCClientDto{
		ClientID:          "confidential-app",
		RedirectURIs:      []string{"http://localhost:3001/callback"},
		TokenEndpointAuth: "client_secret_basic",
	})
	assert.ErrorIs(t, err, errSecretRequired)
}

func TestSPService_UpsertOIDCClient_Confidential_WithSecret(t *testing.T) {
	repo := &upsertOIDCRepo{}
	svc := newTestSPSvc(repo)

	client, err := svc.UpsertOIDCClient(oidcSP(), sp_dto.UpsertOIDCClientDto{
		ClientID:          "confidential-app",
		ClientSecret:      "super-secret",
		RedirectURIs:      []string{"http://localhost:3001/callback"},
		TokenEndpointAuth: "client_secret_basic",
		PKCERequired:      true,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, client.ClientSecretHash)
	assert.True(t, client.PKCERequired)
}

func TestSPService_UpsertOIDCClient_Confidential_UpdatePreservesSecret(t *testing.T) {
	repo := &upsertOIDCRepo{
		oidcClient: &sp_entities.OIDCClient{
			ClientSecretHash: "existing-hash",
		},
	}
	svc := newTestSPSvc(repo)

	client, err := svc.UpsertOIDCClient(oidcSP(), sp_dto.UpsertOIDCClientDto{
		ClientID:          "confidential-app",
		RedirectURIs:      []string{"http://localhost:3001/callback"},
		TokenEndpointAuth: "client_secret_basic",
	})
	require.NoError(t, err)
	assert.Equal(t, "existing-hash", client.ClientSecretHash)
}

func TestSPService_UpsertOIDCClient_Public_RejectsSecret(t *testing.T) {
	repo := &upsertOIDCRepo{}
	svc := newTestSPSvc(repo)

	_, err := svc.UpsertOIDCClient(oidcSP(), sp_dto.UpsertOIDCClientDto{
		ClientID:          "public-spa",
		ClientSecret:      "not-allowed",
		RedirectURIs:      []string{"http://localhost:5173/callback"},
		TokenEndpointAuth: "none",
	})
	assert.ErrorIs(t, err, errSecretNotAllowed)
}
