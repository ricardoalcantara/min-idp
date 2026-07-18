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
	"github.com/ricardoalcantara/min-idp/internal/sp"
	sp_dto "github.com/ricardoalcantara/min-idp/internal/sp/dto"
	"github.com/ricardoalcantara/min-idp/internal/users"
)

const defaultSPSlug = "default"

type BootstrapRepository interface {
	IsInitialized(ctx context.Context) (bool, error)
	SetInitialized(ctx context.Context) error
}

type BootstrapService struct {
	repo          BootstrapRepository
	ks            keystore.KeyStore
	rbacSvc       *rbac.RBACService
	usersSvc      *users.UserService
	spSvc         *sp.SPService
	cfg           *config.Config
	adminPassword string
	log           *slog.Logger
}

func NewBootstrapService(
	repo *bootstrap_repositories.BootstrapRepository,
	ks keystore.KeyStore,
	rbacSvc *rbac.RBACService,
	usersSvc *users.UserService,
	spSvc *sp.SPService,
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
		spSvc:         spSvc,
		cfg:           cfg,
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

	clientID, clientSecret, err := s.createDefaultSP()
	if err != nil {
		return err
	}

	if err := s.repo.SetInitialized(ctx); err != nil {
		return fmt.Errorf("bootstrap: save state: %w", err)
	}

	logArgs := []any{
		"email", adminEmail,
		"password", password,
		"sp_slug", defaultSPSlug,
		"client_id", clientID,
	}
	if s.cfg.BootstrapSPPublic {
		logArgs = append(logArgs, "client_type", "public (PKCE, no secret)")
	} else {
		logArgs = append(logArgs, "client_secret", clientSecret)
	}
	s.log.Info("bootstrap complete — change the admin password immediately", logArgs...)

	return nil
}

func (s *BootstrapService) createDefaultSP() (clientID, clientSecret string, err error) {
	spEntity, err := s.spSvc.Create(defaultSPSlug, s.cfg.BootstrapSPName, "oidc")
	if err != nil {
		return "", "", fmt.Errorf("bootstrap: create default sp: %w", err)
	}

	clientID = s.cfg.BootstrapSPClientID
	if clientID == "" {
		clientID = defaultSPSlug
	}

	redirectURIs := s.cfg.BootstrapSPRedirectURIs
	if len(redirectURIs) == 0 {
		redirectURIs = []string{"http://localhost:3000/callback"}
	}

	input := sp_dto.UpsertOIDCClientDto{
		ClientID:               clientID,
		RedirectURIs:           redirectURIs,
		PostLogoutRedirectURIs: []string{},
	}

	if s.cfg.BootstrapSPPublic {
		input.TokenEndpointAuth = "none"
	} else {
		clientSecret = s.cfg.BootstrapSPClientSecret
		if clientSecret == "" {
			clientSecret, err = localcrypto.GenerateSecureToken(32)
			if err != nil {
				return "", "", fmt.Errorf("bootstrap: generate client secret: %w", err)
			}
		}
		input.ClientSecret = clientSecret
	}

	if _, err = s.spSvc.UpsertOIDCClient(spEntity, input); err != nil {
		return "", "", fmt.Errorf("bootstrap: upsert oidc client: %w", err)
	}

	return clientID, clientSecret, nil
}
