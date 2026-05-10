package sp

import (
	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
	sp_repositories "github.com/ricardoalcantara/min-idp/internal/sp/repositories"
)

type SPService struct {
	repo *sp_repositories.SPRepository
}

func NewSPService(repo *sp_repositories.SPRepository) *SPService {
	return &SPService{repo: repo}
}

func (s *SPService) Create(slug, name, protocol string) (*sp_entities.ServiceProvider, error) {
	sp := &sp_entities.ServiceProvider{
		Slug:     slug,
		Name:     name,
		Protocol: protocol,
		Enabled:  true,
	}
	return sp, s.repo.Create(sp)
}

func (s *SPService) FindByUUID(id string) (*sp_entities.ServiceProvider, error) {
	return s.repo.FindByUUID(id)
}

func (s *SPService) FindBySlug(slug string) (*sp_entities.ServiceProvider, error) {
	return s.repo.FindBySlug(slug)
}

func (s *SPService) List() ([]sp_entities.ServiceProvider, error) {
	return s.repo.FindAll()
}
