package oidc_repositories

import (
	"errors"

	"github.com/go-minstack/repository"
	dbpkg "github.com/ricardoalcantara/min-idp/internal/db"
	oidc_entities "github.com/ricardoalcantara/min-idp/internal/protocol/oidc/entities"
	gormdb "gorm.io/gorm"
)

type OAuthTokenRepository struct {
	*repository.Repository[oidc_entities.OAuthToken]
	db *gormdb.DB
}

func NewOAuthTokenRepository(d *gormdb.DB) *OAuthTokenRepository {
	return &OAuthTokenRepository{
		Repository: repository.NewRepository[oidc_entities.OAuthToken](d),
		db:         d,
	}
}

func (r *OAuthTokenRepository) FindByHash(hash string) (*oidc_entities.OAuthToken, error) {
	t, err := r.FindOne(repository.Where("token_hash = ?", hash))
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, dbpkg.ErrEntityNotFound
	}
	return t, err
}

func (r *OAuthTokenRepository) CreateToken(token *oidc_entities.OAuthToken) error {
	return r.db.Create(token).Error
}

func (r *OAuthTokenRepository) RevokeToken(hash string) error {
	return r.db.Model(&oidc_entities.OAuthToken{}).
		Where("token_hash = ?", hash).
		Update("revoked_at", gormdb.Expr("CURRENT_TIMESTAMP")).Error
}
