package users

import (
	"errors"

	"github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/db"
	user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrAccountNotActive = errors.New("account is not active")

type UserRepository interface {
	Create(u *user_entities.User) error
	FindByID(id uint) (*user_entities.User, error)
	FindByUUID(uuid string) (*user_entities.User, error)
	FindByEmail(email string) (*user_entities.User, error)
	Update(u *user_entities.User) error
	List(page, limit int) ([]*user_entities.User, int64, error)
	Delete(id uint) error
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(email, password string) (*user_entities.User, error) {
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := &user_entities.User{
		Email:        email,
		PasswordHash: hash,
		Status:       "active",
	}
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) FindByID(id uint) (*user_entities.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) FindByUUID(id string) (*user_entities.User, error) {
	return s.repo.FindByUUID(id)
}

func (s *UserService) FindByEmail(email string) (*user_entities.User, error) {
	return s.repo.FindByEmail(email)
}

func (s *UserService) Authenticate(email, password string) (*user_entities.User, error) {
	u, err := s.repo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, db.ErrEntityNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if u.Status != "active" {
		return nil, ErrAccountNotActive
	}
	if err := crypto.VerifyPassword(u.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}
	return u, nil
}

func (s *UserService) List(page, pageSize int) ([]*user_entities.User, int64, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	return s.repo.List(page, pageSize)
}

func (s *UserService) Update(uuid string, email, status *string) (*user_entities.User, error) {
	u, err := s.repo.FindByUUID(uuid)
	if err != nil {
		return nil, err
	}
	if email != nil {
		u.Email = *email
	}
	if status != nil {
		u.Status = *status
	}
	return u, s.repo.Update(u)
}

func (s *UserService) Delete(uuid string) error {
	u, err := s.repo.FindByUUID(uuid)
	if err != nil {
		return err
	}
	return s.repo.Delete(u.ID)
}

func (s *UserService) UpdatePassword(userID uint, newPassword string) error {
	hash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}
	u, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return s.repo.Update(u)
}
