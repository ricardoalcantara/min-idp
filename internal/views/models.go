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
