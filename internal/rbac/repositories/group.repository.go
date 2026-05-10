package rbac_repositories

import (
	"errors"

	"github.com/go-minstack/repository"
	"github.com/ricardoalcantara/min-idp/internal/db"
	rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"
	"gorm.io/gorm"
)

type GroupRepository struct {
	*repository.Repository[rbac_entities.Group]
	db *gorm.DB
}

func NewGroupRepository(d *gorm.DB) *GroupRepository {
	return &GroupRepository{Repository: repository.NewRepository[rbac_entities.Group](d), db: d}
}

func (r *GroupRepository) FindByUUID(uuid string) (*rbac_entities.Group, error) {
	g, err := r.FindOne(repository.Where("uuid = ?", uuid))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return g, err
}

func (r *GroupRepository) AssignToUser(userID, groupID uint) error {
	return r.db.Where("user_id = ? AND group_id = ?", userID, groupID).
		FirstOrCreate(&rbac_entities.UserGroup{UserID: userID, GroupID: groupID}).Error
}

func (r *GroupRepository) RemoveFromUser(userID, groupID uint) error {
	return r.db.Where("user_id = ? AND group_id = ?", userID, groupID).
		Delete(&rbac_entities.UserGroup{}).Error
}

func (r *GroupRepository) GetGroupsByUser(userID uint) ([]rbac_entities.Group, error) {
	var groups []rbac_entities.Group
	err := r.db.
		Joins("JOIN user_groups ON user_groups.group_id = groups.id").
		Where("user_groups.user_id = ?", userID).
		Find(&groups).Error
	return groups, err
}
