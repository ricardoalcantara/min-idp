package rbac_dto

import rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"

type PermissionDto struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewPermissionDto(p *rbac_entities.Permission) PermissionDto {
	return PermissionDto{
		ID:   p.UUID.String(),
		Name: p.Name,
	}
}

type AssignPermissionDto struct {
	Name string `json:"name" binding:"required"`
}
