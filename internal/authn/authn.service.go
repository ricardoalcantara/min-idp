package authn

import (
	"errors"

	"github.com/ricardoalcantara/min-idp/internal/users"
	user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"
)

var errInvalidCredentials = errors.New("invalid credentials")

type UserAuthenticator interface {
	Authenticate(email, password string) (*user_entities.User, error)
}

type AuthnService struct {
	users UserAuthenticator
}

func NewAuthnService(users UserAuthenticator) *AuthnService {
	return &AuthnService{users: users}
}

func (s *AuthnService) Authenticate(email, password string) (*user_entities.User, error) {
	u, err := s.users.Authenticate(email, password)
	if err != nil {
		if errors.Is(err, users.ErrInvalidCredentials) || errors.Is(err, users.ErrAccountNotActive) {
			return nil, errInvalidCredentials
		}
		return nil, err
	}
	return u, nil
}
