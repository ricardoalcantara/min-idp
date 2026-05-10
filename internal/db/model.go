package db

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	gorm.Model
	UUID uuid.UUID `gorm:"uniqueIndex;not null"`
}

func (m *Model) BeforeCreate(_ *gorm.DB) error {
	if m.UUID == uuid.Nil {
		m.UUID = uuid.New()
	}
	return nil
}
