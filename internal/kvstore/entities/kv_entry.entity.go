package kvstore_entities

import "time"

type KVEntry struct {
	Key       string     `gorm:"primaryKey"`
	Value     []byte     `gorm:"not null"`
	ExpiresAt *time.Time `gorm:"index"`
	CreatedAt time.Time
}

func (KVEntry) TableName() string { return "kv_store" }
