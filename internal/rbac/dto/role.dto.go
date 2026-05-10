package rbac_dto

import rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"

type RoleDto struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	System      bool   `json:"system"`
}

func NewRoleDto(r *rbac_entities.Role) RoleDto {
	return RoleDto{
		ID:          r.UUID.String(),
		Name:        r.Name,
		Description: r.Description,
		System:      r.System,
	}
}

type CreateRoleDto struct {
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description"`
}

type UpdateRoleDto struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
