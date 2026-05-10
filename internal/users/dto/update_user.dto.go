package user_dto

type UpdateUserDto struct {
	Email  *string `json:"email"`
	Status *string `json:"status"`
}

type UpdateMeDto struct {
	Email *string `json:"email" binding:"omitempty,email"`
}

type ResetPasswordDto struct {
	Password string `json:"password" binding:"required,min=8"`
}

type AssignRoleDto struct {
	RoleID string `json:"role_id" binding:"required"`
}

type AssignGroupDto struct {
	GroupID string `json:"group_id" binding:"required"`
}
