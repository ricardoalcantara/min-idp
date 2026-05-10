package user_repositories

import (
	"errors"

	"github.com/go-minstack/repository"
	"github.com/ricardoalcantara/min-idp/internal/db"
	user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"
	gormdb "gorm.io/gorm"
)

type UserRepository struct {
	*repository.Repository[user_entities.User]
}

func NewUserRepository(d *gormdb.DB) *UserRepository {
	return &UserRepository{repository.NewRepository[user_entities.User](d)}
}

func (r *UserRepository) FindByEmail(email string) (*user_entities.User, error) {
	u, err := r.FindOne(repository.Where("email = ?", email))
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return u, err
}

func (r *UserRepository) FindByUUID(uuid string) (*user_entities.User, error) {
	u, err := r.FindOne(repository.Where("uuid = ?", uuid))
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return u, err
}
