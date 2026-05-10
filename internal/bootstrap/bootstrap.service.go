package bootstrap

import (
	"context"
	"crypto/elliptic"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
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
	repo      BootstrapRepository
	ks        keystore.KeyStore
	rbacSvc   *rbac.RBACService
	usersSvc  *users.UserService
	masterKey []byte
	log       *slog.Logger
}

func NewBootstrapService(
	repo *bootstrap_repositories.BootstrapRepository,
	ks keystore.KeyStore,
	rbacSvc *rbac.RBACService,
	usersSvc *users.UserService,
	cfg *config.Config,
	log *slog.Logger,
) (*BootstrapService, error) {
	masterKey, err := localcrypto.DecodeMasterKey(cfg.MasterKey)
	if err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}
	return &BootstrapService{
		repo:      repo,
		ks:        ks,
		rbacSvc:   rbacSvc,
		usersSvc:  usersSvc,
		masterKey: masterKey,
		log:       log,
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

	if err := s.createSigningKey(ctx, keystore_entities.ProtocolOIDC, "ES256"); err != nil {
		return fmt.Errorf("bootstrap: oidc key: %w", err)
	}
	if err := s.createSigningKey(ctx, keystore_entities.ProtocolSAML, "RS256"); err != nil {
		return fmt.Errorf("bootstrap: saml key: %w", err)
	}

	for _, name := range []string{"system:admin", "sp:login", "api:user"} {
		role, err := s.rbacSvc.CreateRole(name, "", true)
		if err != nil {
			return fmt.Errorf("bootstrap: role %s: %w", name, err)
		}
		perm, err := s.rbacSvc.CreatePermission(name)
		if err != nil {
			return fmt.Errorf("bootstrap: permission %s: %w", name, err)
		}
		if err := s.rbacSvc.AssignPermissionToRole(role.ID, perm.ID); err != nil {
			return fmt.Errorf("bootstrap: assign permission: %w", err)
		}
	}

	password, err := localcrypto.GenerateSecureToken(12)
	if err != nil {
		return fmt.Errorf("bootstrap: generate password: %w", err)
	}

	adminEmail := "admin@localhost"
	adminUser, err := s.usersSvc.Create(adminEmail, password)
	if err != nil {
		return fmt.Errorf("bootstrap: create admin: %w", err)
	}

	for _, roleName := range []string{"system:admin", "sp:login"} {
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

func (s *BootstrapService) createSigningKey(ctx context.Context, protocol, algorithm string) error {
	var privPEM, pubPEM, certPEM []byte

	switch algorithm {
	case "ES256":
		key, err := localcrypto.GenerateECKey(elliptic.P256())
		if err != nil {
			return err
		}
		privPEM, _ = localcrypto.MarshalPrivateKeyPEM(key)
		pubPEM, _ = localcrypto.MarshalPublicKeyPEM(key.Public())
	case "RS256":
		key, err := localcrypto.GenerateRSAKey(2048)
		if err != nil {
			return err
		}
		privPEM, _ = localcrypto.MarshalPrivateKeyPEM(key)
		pubPEM, _ = localcrypto.MarshalPublicKeyPEM(key.Public())
		if protocol == keystore_entities.ProtocolSAML {
			_, certPEM, _ = localcrypto.GenerateSelfSignedCert(key, "min-idp", 10*365*24*time.Hour)
		}
	}

	encrypted, err := localcrypto.Encrypt(s.masterKey, privPEM)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	return s.ks.InsertKey(ctx, &keystore_entities.SigningKey{
		Protocol:            protocol,
		KID:                 uuid.NewString(),
		Algorithm:           algorithm,
		PrivateKeyEncrypted: encrypted,
		PublicKey:           string(pubPEM),
		Certificate:         string(certPEM),
		Status:              keystore_entities.StatusActive,
		ActivatedAt:         &now,
	})
}
