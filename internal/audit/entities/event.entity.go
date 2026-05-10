package audit_entities

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Timestamp    time.Time `gorm:"index;not null"`
	ActorUserID  *uint
	Action       string `gorm:"index;not null"`
	TargetType   string
	TargetID     *uint
	SPID         *uint
	IP           string
	UserAgent    string
	Result       string `gorm:"default:'ok'"`
	MetadataJSON string
}
