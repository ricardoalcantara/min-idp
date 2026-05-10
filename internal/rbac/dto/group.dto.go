package rbac_dto

import rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"

type GroupDto struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewGroupDto(g *rbac_entities.Group) GroupDto {
	return GroupDto{
		ID:          g.UUID.String(),
		Name:        g.Name,
		Description: g.Description,
	}
}

type CreateGroupDto struct {
	Name        string `json:"name"        binding:"required"`
	Description string `json:"description"`
}

type UpdateGroupDto struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}
