package user_dto

type CreateUserDto struct {
	Email    string `json:"email"    binding:"required,email"`
	Name     string `json:"name"`
	Password string `json:"password" binding:"required,min=8"`
}
