package views

type LoginViewModel struct {
	Next  string
	Error string
}

type InfoViewModel struct {
	Email    string
	Name     string
	UUID     string
	RoleList []string
}

type LogoutViewModel struct {
	SPName    string
	ReturnURL string
	// PublicClient indicates a browser-only client (PKCE, no secret): the IdP
	// cannot re-initiate login on its behalf, so the return link is neutral.
	PublicClient bool
}

type ForgotPasswordViewModel struct {
	Error   string
	Message string
}

type ResetPasswordViewModel struct {
	Token   string
	Error   string
	Message string
}
