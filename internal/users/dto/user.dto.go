package user_dto

import user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"

type UserDto struct {
	UUID   string `json:"uuid"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

func NewUserDto(u *user_entities.User) UserDto {
	return UserDto{
		UUID:   u.UUID.String(),
		Email:  u.Email,
		Status: u.Status,
	}
}
