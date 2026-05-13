package sp

import (
	"errors"
	"testing"

	"github.com/go-minstack/repository"
	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
	"github.com/ricardoalcantara/min-idp/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mocks ---

// mockGateSPRepo implements SPRepository for gate tests.
// Only ListAccessRules is meaningful; other methods return safe defaults.
type mockGateSPRepo struct {
	rules    []sp_repositories.AccessRuleRow
	rulesErr error
}

func (m *mockGateSPRepo) Create(_ *sp_entities.ServiceProvider) error               { return nil }
func (m *mockGateSPRepo) Update(_ *sp_entities.ServiceProvider) error               { return nil }
func (m *mockGateSPRepo) Delete(_ uint) error                                       { return nil }
func (m *mockGateSPRepo) FindByUUID(_ string) (*sp_entities.ServiceProvider, error) { return nil, nil }
func (m *mockGateSPRepo) FindBySlug(_ string) (*sp_entities.ServiceProvider, error) { return nil, nil }
func (m *mockGateSPRepo) FindByID(_ uint) (*sp_entities.ServiceProvider, error) { return nil, nil }
func (m *mockGateSPRepo) FindAll(_ ...repository.QueryOption) ([]sp_entities.ServiceProvider, error) {
	return nil, nil
}
func (m *mockGateSPRepo) GetOIDCClient(_ uint) (*sp_entities.OIDCClient, error)     { return nil, nil }
func (m *mockGateSPRepo) FindOIDCClientByClientID(_ string) (*sp_entities.OIDCClient, error) { return nil, nil }
func (m *mockGateSPRepo) UpsertOIDCClient(_ *sp_entities.OIDCClient) error          { return nil }
func (m *mockGateSPRepo) FindSAMLClientByEntityID(_ string) (*sp_entities.SAMLClient, error) { return nil, nil }
func (m *mockGateSPRepo) GetSAMLClient(_ uint) (*sp_entities.SAMLClient, error)     { return nil, nil }
func (m *mockGateSPRepo) UpsertSAMLClient(_ *sp_entities.SAMLClient) error          { return nil }
func (m *mockGateSPRepo) FindSubjectID(_ string, _ uint) (uint, error)              { return 0, nil }
func (m *mockGateSPRepo) CreateAccessRule(_ *sp_entities.AccessRule) error          { return nil }
func (m *mockGateSPRepo) FindAccessRuleByUUID(_ string) (*sp_entities.AccessRule, error) {
	return nil, nil
}
func (m *mockGateSPRepo) DeleteAccessRule(_ uint) error { return nil }
func (m *mockGateSPRepo) ListAccessRules(_ uint) ([]sp_repositories.AccessRuleRow, error) {
	return m.rules, m.rulesErr
}

// mockGateRBACRepo implements RBACGateRepository for gate tests.
type mockGateRBACRepo struct {
	hasRole    bool
	roleErr    error
	subjectIDs []uint
	subjectErr error
}

func (m *mockGateRBACRepo) UserHasRole(_ uint, _ string) (bool, error) {
	return m.hasRole, m.roleErr
}
func (m *mockGateRBACRepo) GetSubjectIDsForUser(_ uint) ([]uint, error) {
	return m.subjectIDs, m.subjectErr
}

// helpers

func enabledSP(slug string) *sp_entities.ServiceProvider {
	return &sp_entities.ServiceProvider{Slug: slug, Enabled: true, Protocol: types.SPProtocolOIDC}
}

func allowRule(subjectID uint, priority int) sp_repositories.AccessRuleRow {
	return sp_repositories.AccessRuleRow{
		AccessRule: sp_entities.AccessRule{SubjectID: subjectID, RuleType: "allow", Priority: priority},
	}
}

func denyRule(subjectID uint, priority int) sp_repositories.AccessRuleRow {
	return sp_repositories.AccessRuleRow{
		AccessRule: sp_entities.AccessRule{SubjectID: subjectID, RuleType: "deny", Priority: priority},
	}
}

func newGateSvc(spRepo SPRepository, rbacRepo RBACGateRepository) *SPGateService {
	return NewSPGateService(spRepo, rbacRepo)
}

// --- tests ---

func TestSPGateService_CanSSO_SPDisabled(t *testing.T) {
	sp := &sp_entities.ServiceProvider{Slug: "app", Enabled: false}
	svc := newGateSvc(&mockGateSPRepo{}, &mockGateRBACRepo{hasRole: true})

	err := svc.CanSSO(1, sp)
	assert.ErrorIs(t, err, ErrSPDisabled)
}

func TestSPGateService_CanSSO_NoGlobalPermission(t *testing.T) {
	svc := newGateSvc(&mockGateSPRepo{}, &mockGateRBACRepo{hasRole: false})

	err := svc.CanSSO(1, enabledSP("app"))
	assert.ErrorIs(t, err, ErrAccessDenied)
}

func TestSPGateService_CanSSO_GlobalPermissionChecksError(t *testing.T) {
	dbErr := errors.New("db down")
	svc := newGateSvc(&mockGateSPRepo{}, &mockGateRBACRepo{roleErr: dbErr})

	err := svc.CanSSO(1, enabledSP("app"))
	assert.ErrorIs(t, err, dbErr)
}

func TestSPGateService_CanSSO_NoRulesDefaultAllow(t *testing.T) {
	// User has sp:login and SP has no access rules → default allow (minimum effort setup)
	svc := newGateSvc(
		&mockGateSPRepo{rules: []sp_repositories.AccessRuleRow{}},
		&mockGateRBACRepo{hasRole: true, subjectIDs: []uint{10}},
	)

	err := svc.CanSSO(1, enabledSP("app"))
	assert.NoError(t, err)
}

func TestSPGateService_CanSSO_AllowRuleMatchesRole(t *testing.T) {
	// Subject 10 = user's role-subject; rule allows it
	svc := newGateSvc(
		&mockGateSPRepo{rules: []sp_repositories.AccessRuleRow{allowRule(10, 0)}},
		&mockGateRBACRepo{hasRole: true, subjectIDs: []uint{10}},
	)

	err := svc.CanSSO(1, enabledSP("app"))
	require.NoError(t, err)
}

func TestSPGateService_CanSSO_DenyRuleBeforeAllow(t *testing.T) {
	// deny(priority=0) comes before allow(priority=1); deny wins for subject 10
	svc := newGateSvc(
		&mockGateSPRepo{rules: []sp_repositories.AccessRuleRow{
			denyRule(10, 0),
			allowRule(10, 1),
		}},
		&mockGateRBACRepo{hasRole: true, subjectIDs: []uint{10}},
	)

	err := svc.CanSSO(1, enabledSP("app"))
	assert.ErrorIs(t, err, ErrAccessDenied)
}

func TestSPGateService_CanSSO_NoMatchingRuleDefaultDeny(t *testing.T) {
	// Rules exist but none match the user's subjects
	svc := newGateSvc(
		&mockGateSPRepo{rules: []sp_repositories.AccessRuleRow{allowRule(99, 0)}},
		&mockGateRBACRepo{hasRole: true, subjectIDs: []uint{10, 20}},
	)

	err := svc.CanSSO(1, enabledSP("app"))
	assert.ErrorIs(t, err, ErrAccessDenied)
}

func TestSPGateService_CanSSO_FineGrainedSlugPermission(t *testing.T) {
	// No global sp:login, but has sp:login:app — gate should accept it
	callCount := 0
	rbac := &stubSlugRBACRepo{slug: "app", callCount: &callCount}
	svc := newGateSvc(
		&mockGateSPRepo{rules: []sp_repositories.AccessRuleRow{allowRule(10, 0)}},
		rbac,
	)

	err := svc.CanSSO(1, enabledSP("app"))
	require.NoError(t, err)
}

// stubSlugRBACRepo returns false for "sp:login" but true for "sp:login:<slug>"
type stubSlugRBACRepo struct {
	slug      string
	callCount *int
}

func (s *stubSlugRBACRepo) UserHasRole(_ uint, perm string) (bool, error) {
	if perm == "sp:login:"+s.slug {
		return true, nil
	}
	return false, nil
}
func (s *stubSlugRBACRepo) GetSubjectIDsForUser(_ uint) ([]uint, error) {
	return []uint{10}, nil
}
