package oidc_dto

// UserInfoResponse represents the standard OpenID Connect UserInfo response.
type UserInfoResponse struct {
	Sub      string   `json:"sub"`
	Email    string   `json:"email,omitempty"`
	Username string   `json:"username,omitempty"`
	Name     string   `json:"name,omitempty"`
	Roles             []string `json:"roles,omitempty"`
	UpdatedAt         int64    `json:"updated_at,omitempty"`
}
