package keystore_dto

import (
	"time"

	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
)

type KeyDto struct {
	KID         string     `json:"kid"`
	Protocol    string     `json:"protocol"`
	Algorithm   string     `json:"algorithm"`
	Status      string     `json:"status"`
	PublicKey   string     `json:"public_key"`
	Certificate string     `json:"certificate,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	RetiredAt   *time.Time `json:"retired_at,omitempty"`
}

func NewKeyDto(k *keystore_entities.SigningKey) KeyDto {
	return KeyDto{
		KID:         k.KID,
		Protocol:    k.Protocol,
		Algorithm:   k.Algorithm,
		Status:      k.Status,
		PublicKey:   k.PublicKey,
		Certificate: k.Certificate,
		CreatedAt:   k.CreatedAt,
		ActivatedAt: k.ActivatedAt,
		RetiredAt:   k.RetiredAt,
	}
}
