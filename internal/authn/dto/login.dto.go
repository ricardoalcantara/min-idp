package authn_dto

type LoginDto struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}
