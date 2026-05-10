package keystore_entities

import (
	"time"

	"gorm.io/gorm"
)

const (
	StatusActive   = "active"
	StatusPrevious = "previous"
	StatusRetired  = "retired"

	ProtocolOIDC = "oidc"
	ProtocolSAML = "saml"
)

type SigningKey struct {
	gorm.Model
	Protocol            string     `gorm:"not null;index:idx_key_proto_status"`
	KID                 string     `gorm:"column:kid;not null"`
	Algorithm           string     `gorm:"not null"`
	PrivateKeyEncrypted []byte     `gorm:"not null"`
	PublicKey           string     `gorm:"not null"`
	Certificate         string
	Status              string     `gorm:"not null;default:'active';index:idx_key_proto_status"`
	ActivatedAt         *time.Time
	RetiredAt           *time.Time
}
