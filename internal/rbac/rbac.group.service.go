package rbac

import (
	"github.com/go-minstack/repository"
	rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"
	rbac_repositories "github.com/ricardoalcantara/min-idp/internal/rbac/repositories"
)

type GroupRepository interface {
	Create(g *rbac_entities.Group) error
	FindByID(id uint) (*rbac_entities.Group, error)
	FindByUUID(uuid string) (*rbac_entities.Group, error)
	FindAll(opts ...repository.QueryOption) ([]rbac_entities.Group, error)
	Update(g *rbac_entities.Group) error
	Delete(g *rbac_entities.Group) error
	AssignToUser(userID, groupID uint) error
	RemoveFromUser(userID, groupID uint) error
	GetGroupsByUser(userID uint) ([]rbac_entities.Group, error)
}

type GroupService struct {
	repo GroupRepository
}

func NewGroupService(repo *rbac_repositories.GroupRepository) *GroupService {
	return &GroupService{repo: repo}
}

func (s *GroupService) Create(name, description string) (*rbac_entities.Group, error) {
	g := &rbac_entities.Group{Name: name, Description: description}
	return g, s.repo.Create(g)
}

func (s *GroupService) FindByUUID(uuid string) (*rbac_entities.Group, error) {
	return s.repo.FindByUUID(uuid)
}

func (s *GroupService) List() ([]rbac_entities.Group, error) {
	return s.repo.FindAll()
}

func (s *GroupService) Update(uuid string, name, description *string) (*rbac_entities.Group, error) {
	g, err := s.repo.FindByUUID(uuid)
	if err != nil {
		return nil, err
	}
	if name != nil {
		g.Name = *name
	}
	if description != nil {
		g.Description = *description
	}
	return g, s.repo.Update(g)
}

func (s *GroupService) Delete(uuid string) error {
	g, err := s.repo.FindByUUID(uuid)
	if err != nil {
		return err
	}
	return s.repo.Delete(g)
}

func (s *GroupService) AssignToUserByUUID(userID uint, groupUUID string) error {
	g, err := s.repo.FindByUUID(groupUUID)
	if err != nil {
		return err
	}
	return s.repo.AssignToUser(userID, g.ID)
}

func (s *GroupService) RemoveFromUserByUUID(userID uint, groupUUID string) error {
	g, err := s.repo.FindByUUID(groupUUID)
	if err != nil {
		return err
	}
	return s.repo.RemoveFromUser(userID, g.ID)
}

func (s *GroupService) GetGroupsByUser(userID uint) ([]rbac_entities.Group, error) {
	return s.repo.GetGroupsByUser(userID)
}
