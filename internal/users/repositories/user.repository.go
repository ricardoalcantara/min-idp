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
	db *gormdb.DB
}

func NewUserRepository(d *gormdb.DB) *UserRepository {
	return &UserRepository{Repository: repository.NewRepository[user_entities.User](d), db: d}
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

func (r *UserRepository) List(offset, limit int) ([]*user_entities.User, int64, error) {
	var users []*user_entities.User
	var total int64
	r.db.Model(&user_entities.User{}).Count(&total)
	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&user_entities.User{}, id).Error
}
