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

func (r *UserRepository) Create(u *user_entities.User) error {
	return r.db.Transaction(func(tx *gormdb.DB) error {
		if err := tx.Create(u).Error; err != nil {
			return err
		}
		return tx.Create(&db.Subject{Type: db.SubjectTypeUser, EntityID: u.ID}).Error
	})
}

func (r *UserRepository) FindByEmail(email string) (*user_entities.User, error) {
	var u user_entities.User
	err := r.db.Where("email = ?", email).First(&u).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return &u, err
}

func (r *UserRepository) FindByUUID(uuid string) (*user_entities.User, error) {
	var u user_entities.User
	err := r.db.Preload("Roles").Where("uuid = ?", uuid).First(&u).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return &u, err
}

func (r *UserRepository) FindByID(id uint) (*user_entities.User, error) {
	var u user_entities.User
	err := r.db.Preload("Roles").First(&u, id).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return &u, err
}

func (r *UserRepository) List(page, limit int) ([]*user_entities.User, int64, error) {
	total, err := r.Count()
	if err != nil {
		return nil, 0, err
	}
	users, err := r.FindAll(
		repository.Preload("Roles"),
		repository.Paginate(repository.NewPagination(page, limit)),
	)
	if err != nil {
		return nil, 0, err
	}
	ptrs := make([]*user_entities.User, len(users))
	for i := range users {
		ptrs[i] = &users[i]
	}
	return ptrs, total, nil
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&user_entities.User{}, id).Error
}
