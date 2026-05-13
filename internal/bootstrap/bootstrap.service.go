package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	bootstrap_repositories "github.com/ricardoalcantara/min-idp/internal/bootstrap/repositories"
	"github.com/ricardoalcantara/min-idp/internal/config"
	localcrypto "github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/keystore"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	"github.com/ricardoalcantara/min-idp/internal/rbac"
	"github.com/ricardoalcantara/min-idp/internal/users"
)

type BootstrapRepository interface {
	IsInitialized(ctx context.Context) (bool, error)
	SetInitialized(ctx context.Context) error
}

type BootstrapService struct {
	repo          BootstrapRepository
	ks            keystore.KeyStore
	rbacSvc       *rbac.RBACService
	usersSvc      *users.UserService
	adminPassword string
	log           *slog.Logger
}

func NewBootstrapService(
	repo *bootstrap_repositories.BootstrapRepository,
	ks keystore.KeyStore,
	rbacSvc *rbac.RBACService,
	usersSvc *users.UserService,
	cfg *config.Config,
	log *slog.Logger,
) (*BootstrapService, error) {
	if _, err := localcrypto.DecodeMasterKey(cfg.MasterKey); err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	return &BootstrapService{
		repo:          repo,
		ks:            ks,
		rbacSvc:       rbacSvc,
		usersSvc:      usersSvc,
		adminPassword: cfg.AdminPassword,
		log:           log,
	}, nil
}

func (s *BootstrapService) Run(ctx context.Context) error {
	initialized, err := s.repo.IsInitialized(ctx)
	if err != nil {
		return fmt.Errorf("bootstrap: check state: %w", err)
	}
	if initialized {
		s.log.Info("bootstrap: already initialized")
		return nil
	}

	s.log.Info("bootstrap: first run, initializing...")

	if err := s.ks.GenerateAndRotate(ctx, keystore_entities.ProtocolOIDC); err != nil {
		return fmt.Errorf("bootstrap: oidc key: %w", err)
	}
	if err := s.ks.GenerateAndRotate(ctx, keystore_entities.ProtocolSAML); err != nil {
		return fmt.Errorf("bootstrap: saml key: %w", err)
	}

	systemRoles := []string{"system:admin", "sp:login", "api:user"}
	for _, name := range systemRoles {
		if _, err := s.rbacSvc.CreateRole(name, "", true); err != nil {
			return fmt.Errorf("bootstrap: role %s: %w", name, err)
		}
	}

	password := s.adminPassword
	if password == "" {
		password, err = localcrypto.GenerateSecureToken(12)
		if err != nil {
			return fmt.Errorf("bootstrap: generate password: %w", err)
		}
	}

	adminEmail := "admin@min-idp.local"
	adminUser, err := s.usersSvc.Create(adminEmail, "admin", "Admin", password)
	if err != nil {
		return fmt.Errorf("bootstrap: create admin: %w", err)
	}

	for _, roleName := range systemRoles {
		role, err := s.rbacSvc.FindRoleByName(roleName)
		if err != nil {
			return fmt.Errorf("bootstrap: find role %s: %w", roleName, err)
		}
		if err := s.rbacSvc.AssignRoleToUser(adminUser.ID, role.ID); err != nil {
			return fmt.Errorf("bootstrap: assign role: %w", err)
		}
	}

	if err := s.repo.SetInitialized(ctx); err != nil {
		return fmt.Errorf("bootstrap: save state: %w", err)
	}

	s.log.Info("bootstrap complete — change the admin password immediately",
		"email", adminEmail,
		"password", password,
	)

	return nil
}

