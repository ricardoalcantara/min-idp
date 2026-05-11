package user_dto

import user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"

type UserDto struct {
	UUID   string   `json:"uuid"`
	Email  string   `json:"email"`
	Status string   `json:"status"`
	Roles  []string `json:"roles"`
}

func NewUserDto(u *user_entities.User) UserDto {
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = r.Name
	}
	return UserDto{
		UUID:   u.UUID.String(),
		Email:  u.Email,
		Status: u.Status,
		Roles:  roles,
	}
}
