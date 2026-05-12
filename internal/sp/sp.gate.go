package sp

import (
	"errors"

	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
)

// ErrAccessDenied is returned when a user is not permitted to SSO into an SP.
// Exported so Tier 5/6 controllers can map it to the right HTTP response (403).
var ErrAccessDenied = errors.New("access denied")

// ErrSPDisabled is returned when the SP is disabled.
var ErrSPDisabled = errors.New("service provider is disabled")

// RBACGateRepository is the minimal RBAC interface the SSO gate needs.
// Defined here so the sp package has no import dependency on the rbac package.
type RBACGateRepository interface {
	UserHasPermission(userID uint, permission string) (bool, error)
	GetSubjectIDsForUser(userID uint) ([]uint, error)
}

// SPGateService evaluates whether a user is allowed to SSO into a given SP.
// It implements the two-step algorithm from spec §7:
//  1. Global sp:login (or sp:login:<slug>) permission check.
//  2. Per-SP access_rules walk in priority order; default deny if no allow matches.
type SPGateService struct {
	spRepo   SPRepository
	rbacRepo RBACGateRepository
}

func NewSPGateService(spRepo SPRepository, rbacRepo RBACGateRepository) *SPGateService {
	return &SPGateService{spRepo: spRepo, rbacRepo: rbacRepo}
}

// CanSSO returns nil if the user is allowed to SSO into sp, or a sentinel error otherwise.
func (g *SPGateService) CanSSO(userID uint, sp *sp_entities.ServiceProvider) error {
	if !sp.Enabled {
		return ErrSPDisabled
	}

	// Step 1 — global gate: sp:login OR sp:login:<slug>
	hasGlobal, err := g.rbacRepo.UserHasPermission(userID, "sp:login")
	if err != nil {
		return err
	}
	if !hasGlobal {
		hasSlug, err := g.rbacRepo.UserHasPermission(userID, "sp:login:"+sp.Slug)
		if err != nil {
			return err
		}
		if !hasSlug {
			return ErrAccessDenied
		}
	}

	// Step 2 — per-SP rule walk (rules come back ORDER BY priority ASC)
	rules, err := g.spRepo.ListAccessRules(sp.ID)
	if err != nil {
		return err
	}

	// No rules configured → default allow for sp:login holders (minimum effort setup)
	if len(rules) == 0 {
		return nil
	}

	// Collect the set of subjects.id values this user belongs to
	subjectIDs, err := g.rbacRepo.GetSubjectIDsForUser(userID)
	if err != nil {
		return err
	}
	subjectSet := make(map[uint]struct{}, len(subjectIDs))
	for _, id := range subjectIDs {
		subjectSet[id] = struct{}{}
	}

	// First matching rule wins
	for _, rule := range rules {
		if _, matched := subjectSet[rule.SubjectID]; matched {
			if rule.RuleType == "allow" {
				return nil
			}
			return ErrAccessDenied
		}
	}

	// No rule matched → default deny
	return ErrAccessDenied
}
