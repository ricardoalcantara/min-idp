package authn_dto

import "github.com/google/uuid"

type LoginResponseDto struct {
	SessionID uuid.UUID `json:"session_id"`
}
