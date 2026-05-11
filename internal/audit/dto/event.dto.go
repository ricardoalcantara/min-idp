package audit_dto

import (
	"time"

	audit_entities "github.com/ricardoalcantara/min-idp/internal/audit/entities"
)

type EventDto struct {
	ID           uint      `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	ActorUserID  *uint     `json:"actor_user_id,omitempty"`
	Action       string    `json:"action"`
	TargetType   string    `json:"target_type,omitempty"`
	TargetID     *uint     `json:"target_id,omitempty"`
	SPID         *uint     `json:"sp_id,omitempty"`
	IP           string    `json:"ip,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	Result       string    `json:"result"`
	MetadataJSON string    `json:"metadata,omitempty"`
}

func NewEventDto(e *audit_entities.Event) EventDto {
	return EventDto{
		ID:           e.ID,
		Timestamp:    e.Timestamp,
		ActorUserID:  e.ActorUserID,
		Action:       e.Action,
		TargetType:   e.TargetType,
		TargetID:     e.TargetID,
		SPID:         e.SPID,
		IP:           e.IP,
		UserAgent:    e.UserAgent,
		Result:       e.Result,
		MetadataJSON: e.MetadataJSON,
	}
}

type EventsListDto struct {
	Data     []EventDto `json:"data"`
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}
