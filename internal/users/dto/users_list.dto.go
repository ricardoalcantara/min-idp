package user_dto

type UsersListDto struct {
	Data     []UserDto `json:"data"`
	Total    int64     `json:"total"`
	Page     int       `json:"page"`
	PageSize int       `json:"page_size"`
}
