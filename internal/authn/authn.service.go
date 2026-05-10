package authn

import (
	"errors"

	"github.com/ricardoalcantara/min-idp/internal/users"
	user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"
)

var errInvalidCredentials = errors.New("invalid credentials")

type AuthnService struct {
	users *users.UserService
}

func NewAuthnService(users *users.UserService) *AuthnService {
	return &AuthnService{users: users}
}

func (s *AuthnService) Authenticate(email, password string) (*user_entities.User, error) {
	u, err := s.users.Authenticate(email, password)
	if err != nil {
		return nil, errInvalidCredentials
	}
	return u, nil
}
