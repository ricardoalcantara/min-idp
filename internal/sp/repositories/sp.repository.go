package sp_repositories

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-minstack/go-minstack/repository"
	dbpkg "github.com/ricardoalcantara/min-idp/internal/db"
	sp_entities "github.com/ricardoalcantara/min-idp/internal/sp/entities"
	gormdb "gorm.io/gorm"
)

type SPRepository struct {
	*repository.Repository[sp_entities.ServiceProvider]
	db *gormdb.DB
}

func NewSPRepository(d *gormdb.DB) *SPRepository {
	return &SPRepository{Repository: repository.NewRepository[sp_entities.ServiceProvider](d), db: d}
}

func (r *SPRepository) FindByUUID(uuid string) (*sp_entities.ServiceProvider, error) {
	s, err := r.FindOne(repository.Where("uuid = ?", uuid))
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, dbpkg.ErrEntityNotFound
	}
	return s, err
}

func (r *SPRepository) FindByID(id uint) (*sp_entities.ServiceProvider, error) {
	s, err := r.FindOne(repository.Where("id = ?", id))
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, dbpkg.ErrEntityNotFound
	}
	return s, err
}

func (r *SPRepository) FindBySlug(slug string) (*sp_entities.ServiceProvider, error) {
	s, err := r.FindOne(repository.Where("slug = ?", slug))
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, dbpkg.ErrEntityNotFound
	}
	return s, err
}

func (r *SPRepository) Update(sp *sp_entities.ServiceProvider) error {
	return r.db.Save(sp).Error
}

func (r *SPRepository) Delete(id uint) error {
	return r.db.Delete(&sp_entities.ServiceProvider{}, id).Error
}

// --- OIDC Client ---

func (r *SPRepository) FindOIDCClientByClientID(clientID string) (*sp_entities.OIDCClient, error) {
	var c sp_entities.OIDCClient
	err := r.db.Where("client_id = ?", clientID).First(&c).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, dbpkg.ErrEntityNotFound
	}
	return &c, err
}

func (r *SPRepository) GetOIDCClient(spID uint) (*sp_entities.OIDCClient, error) {
	var c sp_entities.OIDCClient
	err := r.db.Where("sp_id = ?", spID).First(&c).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, dbpkg.ErrEntityNotFound
	}
	return &c, err
}

func (r *SPRepository) UpsertOIDCClient(c *sp_entities.OIDCClient) error {
	var existing sp_entities.OIDCClient
	err := r.db.Where("sp_id = ?", c.SPID).First(&existing).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return r.db.Create(c).Error
	}
	if err != nil {
		return err
	}
	c.ID = existing.ID
	c.CreatedAt = existing.CreatedAt
	return r.db.Save(c).Error
}

// --- SAML Client ---

func (r *SPRepository) GetSAMLClient(spID uint) (*sp_entities.SAMLClient, error) {
	var c sp_entities.SAMLClient
	err := r.db.Where("sp_id = ?", spID).First(&c).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, dbpkg.ErrEntityNotFound
	}
	return &c, err
}

func (r *SPRepository) FindSAMLClientByEntityID(entityID string) (*sp_entities.SAMLClient, error) {
	var c sp_entities.SAMLClient
	err := r.db.Where("entity_id = ?", entityID).First(&c).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, dbpkg.ErrEntityNotFound
	}
	return &c, err
}

func (r *SPRepository) UpsertSAMLClient(c *sp_entities.SAMLClient) error {
	var existing sp_entities.SAMLClient
	err := r.db.Where("sp_id = ?", c.SPID).First(&existing).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return r.db.Create(c).Error
	}
	if err != nil {
		return err
	}
	c.ID = existing.ID
	c.CreatedAt = existing.CreatedAt
	return r.db.Save(c).Error
}

// --- Access Rules ---

// AccessRuleRow is used for the JOIN query that resolves subject UUID.
type AccessRuleRow struct {
	sp_entities.AccessRule
	SubjectType string
	SubjectUUID string
}

func (r *SPRepository) ListAccessRules(spID uint) ([]AccessRuleRow, error) {
	rows, err := r.db.Raw(`
		SELECT ar.*, s.type as subject_type,
		  CASE s.type
		    WHEN 'role'  THEN r.uuid
		    
		    WHEN 'user'  THEN u.uuid
		  END as subject_uuid
		FROM access_rules ar
		JOIN subjects s ON ar.subject_id = s.id
		LEFT JOIN roles r ON s.type = 'role'  AND s.entity_id = r.id
		LEFT JOIN users u ON s.type = 'user'  AND s.entity_id = u.id
		WHERE ar.sp_id = ? AND ar.deleted_at IS NULL
		ORDER BY ar.priority ASC`, spID).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AccessRuleRow
	for rows.Next() {
		var row AccessRuleRow
		if err := r.db.ScanRows(rows, &row); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil
}

func (r *SPRepository) FindSubjectID(subjectType string, entityID uint) (uint, error) {
	var s dbpkg.Subject
	err := r.db.Where("type = ? AND entity_id = ?", subjectType, entityID).First(&s).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return 0, fmt.Errorf("subject not found for %s id %d", subjectType, entityID)
	}
	return s.ID, err
}

func (r *SPRepository) CreateAccessRule(rule *sp_entities.AccessRule) error {
	return r.db.Create(rule).Error
}

func (r *SPRepository) FindAccessRuleByUUID(uuid string) (*sp_entities.AccessRule, error) {
	var rule sp_entities.AccessRule
	err := r.db.Where("uuid = ?", uuid).First(&rule).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, dbpkg.ErrEntityNotFound
	}
	return &rule, err
}

func (r *SPRepository) DeleteAccessRule(id uint) error {
	return r.db.Delete(&sp_entities.AccessRule{}, id).Error
}

// Helpers for JSON array ↔ []string

func MarshalStringSlice(v []string) string {
	if v == nil {
		v = []string{}
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func UnmarshalStringSlice(s string) []string {
	var v []string
	_ = json.Unmarshal([]byte(s), &v)
	return v
}
