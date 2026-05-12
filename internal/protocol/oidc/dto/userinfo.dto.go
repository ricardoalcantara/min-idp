package oidc_dto

// UserInfoResponse represents the standard OpenID Connect UserInfo response.
type UserInfoResponse struct {
	Sub         string `json:"sub"`
	Email       string `json:"email,omitempty"`
	Roles       string `json:"roles,omitempty"` // Example custom claim for min-idp roles
	UpdatedAt   int64  `json:"updated_at,omitempty"`
}
