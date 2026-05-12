package oidc_dto

// OAuth2ErrorDto follows RFC 6749 §5.2 error response format.
type OAuth2ErrorDto struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

func NewOAuth2Error(code, description string) OAuth2ErrorDto {
	return OAuth2ErrorDto{Error: code, ErrorDescription: description}
}

// TokenRequest represents a request to the /oauth2/token endpoint.
// Modeled for parsing application/x-www-form-urlencoded data.
type TokenRequest struct {
	GrantType    string `form:"grant_type" binding:"required"`
	Code         string `form:"code"`
	RedirectURI  string `form:"redirect_uri"`
	ClientID     string `form:"client_id"`
	ClientSecret string `form:"client_secret"`
	CodeVerifier string `form:"code_verifier"`
	RefreshToken string `form:"refresh_token"`
	Scope        string `form:"scope"`
}

// TokenResponse represents a successful token exchange.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // in seconds
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
}

type IntrospectRequest struct {
	Token string `form:"token" binding:"required"`
}

type IntrospectResponse struct {
	Active    bool   `json:"active"`
	Scope     string `json:"scope,omitempty"`
	ClientID  string `json:"client_id,omitempty"`
	Username  string `json:"username,omitempty"`
	TokenType string `json:"token_type,omitempty"`
	Exp       int64  `json:"exp,omitempty"`
	Iat       int64  `json:"iat,omitempty"`
	Sub       string `json:"sub,omitempty"`
}

type RevokeRequest struct {
	Token string `form:"token" binding:"required"`
}
