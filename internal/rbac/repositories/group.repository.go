package rbac_repositories

import (
	"github.com/go-minstack/repository"
	rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"
	"gorm.io/gorm"
)

type GroupRepository struct {
	*repository.Repository[rbac_entities.Group]
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{repository.NewRepository[rbac_entities.Group](db)}
}
