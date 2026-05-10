package users

import (
	"errors"

	"github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/db"
	user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"
	user_repositories "github.com/ricardoalcantara/min-idp/internal/users/repositories"
)

var errInvalidCredentials = errors.New("invalid credentials")
var errAccountNotActive = errors.New("account is not active")

type UserService struct {
	users *user_repositories.UserRepository
}

func NewUserService(users *user_repositories.UserRepository) *UserService {
	return &UserService{users: users}
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
	if err := s.users.Create(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) FindByID(id uint) (*user_entities.User, error) {
	return s.users.FindByID(id)
}

func (s *UserService) FindByUUID(id string) (*user_entities.User, error) {
	return s.users.FindByUUID(id)
}

func (s *UserService) FindByEmail(email string) (*user_entities.User, error) {
	return s.users.FindByEmail(email)
}

func (s *UserService) Authenticate(email, password string) (*user_entities.User, error) {
	u, err := s.users.FindByEmail(email)
	if err != nil {
		if err == db.ErrEntityNotFound {
			return nil, errInvalidCredentials
		}
		return nil, err
	}
	if u.Status != "active" {
		return nil, errAccountNotActive
	}
	if err := crypto.VerifyPassword(u.PasswordHash, password); err != nil {
		return nil, errInvalidCredentials
	}
	return u, nil
}

func (s *UserService) UpdatePassword(userID uint, newPassword string) error {
	hash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return err
	}
	u, err := s.users.FindByID(userID)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return s.users.Update(u)
}
