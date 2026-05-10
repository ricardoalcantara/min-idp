package sp_repositories

import (
	"errors"

	"github.com/go-minstack/repository"
	"github.com/ricardoalcantara/min-idp/internal/db"
	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
	gormdb "gorm.io/gorm"
)

type SPRepository struct {
	*repository.Repository[sp_entities.ServiceProvider]
	db *gormdb.DB
}

func NewSPRepository(d *gormdb.DB) *SPRepository {
	return &SPRepository{Repository: repository.NewRepository[sp_entities.ServiceProvider](d), db: d}
}

func (r *SPRepository) FindByUUID(uuid string) (*sp_entities.ServiceProvider, error) {
	s, err := r.FindOne(repository.Where("uuid = ?", uuid))
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return s, err
}

func (r *SPRepository) FindBySlug(slug string) (*sp_entities.ServiceProvider, error) {
	s, err := r.FindOne(repository.Where("slug = ?", slug))
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return s, err
}

func (r *SPRepository) GetOIDCClientByClientID(clientID string) (*sp_entities.OIDCClient, error) {
	var c sp_entities.OIDCClient
	err := r.db.Where("client_id = ?", clientID).First(&c).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return &c, err
}

func (r *SPRepository) GetSAMLClientByEntityID(entityID string) (*sp_entities.SAMLClient, error) {
	var c sp_entities.SAMLClient
	err := r.db.Where("entity_id = ?", entityID).First(&c).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return &c, err
}
