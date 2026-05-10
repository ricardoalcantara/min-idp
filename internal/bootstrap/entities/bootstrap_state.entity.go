package bootstrap_entities

type BootstrapState struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"not null"`
}
