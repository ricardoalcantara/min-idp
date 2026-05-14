package user_dto

import user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"

type UserDto struct {
	UUID     string   `json:"uuid"`
	Email    string   `json:"email"`
	Username string   `json:"username,omitempty"`
	Name     string   `json:"name,omitempty"`
	Status   string   `json:"status"`
	Roles    []string `json:"roles"`
}

func NewUserDto(u *user_entities.User) UserDto {
	roles := make([]string, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = r.Name
	}
	return UserDto{
		UUID:     u.UUID.String(),
		Email:    u.Email,
		Username: u.Username,
		Name:     u.Name,
		Status:   string(u.Status),
		Roles:    roles,
	}
}
