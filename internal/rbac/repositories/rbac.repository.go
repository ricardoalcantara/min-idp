package rbac_repositories

import (
	"errors"

	"github.com/go-minstack/go-minstack/repository"
	"github.com/ricardoalcantara/min-idp/internal/db"
	rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"
	"gorm.io/gorm"
)

type RBACRepository struct {
	*repository.Repository[rbac_entities.Role]
	db *gorm.DB
}

func NewRBACRepository(d *gorm.DB) *RBACRepository {
	return &RBACRepository{Repository: repository.NewRepository[rbac_entities.Role](d), db: d}
}

func (r *RBACRepository) CreateRole(role *rbac_entities.Role) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}
		return tx.Create(&db.Subject{Type: db.SubjectTypeRole, EntityID: role.ID}).Error
	})
}

func (r *RBACRepository) FindRoleByName(name string) (*rbac_entities.Role, error) {
	role, err := r.FindOne(repository.Where("name = ?", name))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return role, err
}

func (r *RBACRepository) FindRoleByID(id uint) (*rbac_entities.Role, error) {
	role, err := r.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return role, err
}

func (r *RBACRepository) FindRoleByUUID(uuid string) (*rbac_entities.Role, error) {
	role, err := r.FindOne(repository.Where("uuid = ?", uuid))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return role, err
}

func (r *RBACRepository) ListRoles() ([]rbac_entities.Role, error) {
	return r.FindAll()
}

func (r *RBACRepository) UpdateRole(role *rbac_entities.Role) error {
	return r.Update(role)
}

func (r *RBACRepository) DeleteRole(id uint) error {
	return r.DeleteByID(id)
}

func (r *RBACRepository) AssignRoleToUser(userID, roleID uint) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		FirstOrCreate(&rbac_entities.UserRole{UserID: userID, RoleID: roleID}).Error
}

func (r *RBACRepository) RemoveRoleFromUser(userID, roleID uint) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&rbac_entities.UserRole{}).Error
}

func (r *RBACRepository) GetRolesByUser(userID uint) ([]rbac_entities.Role, error) {
	var roles []rbac_entities.Role
	err := r.db.
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

func (r *RBACRepository) UserHasRole(userID uint, role string) (bool, error) {
	var count int64
	err := r.db.Model(&rbac_entities.Role{}).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ? AND roles.name = ?", userID, role).
		Count(&count).Error
	return count > 0, err
}

func (r *RBACRepository) GetUserRoleNames(userID uint) ([]string, error) {
	var names []string
	err := r.db.Model(&rbac_entities.Role{}).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Pluck("roles.name", &names).Error
	return names, err
}

// GetSubjectIDsForUser returns the subjects.id values for every principal the
// user belongs to: their own user-subject and all role-subjects.
// Used by the SSO gate to match access_rules.subject_id in one query.
func (r *RBACRepository) GetSubjectIDsForUser(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Raw(`
		SELECT s.id FROM subjects s
		WHERE (s.type = 'user' AND s.entity_id = ?)
		   OR (s.type = 'role' AND s.entity_id IN (
		           SELECT role_id FROM user_roles WHERE user_id = ?
		       ))
	`, userID, userID).Scan(&ids).Error
	return ids, err
}
