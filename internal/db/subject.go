package db

const (
	SubjectTypeRole = "role"
	SubjectTypeUser = "user"
)

// Subject is the unified principals table. Roles and users each get
// an entry here when created so access_rules can reference them with a single FK.
type Subject struct {
	ID       uint   `gorm:"primaryKey"`
	Type     string `gorm:"not null;uniqueIndex:idx_subject"`
	EntityID uint   `gorm:"not null;uniqueIndex:idx_subject"`
}
