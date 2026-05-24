package notification_dto

type SendTestDto struct {
	To string `json:"to" binding:"required,email"`
}
