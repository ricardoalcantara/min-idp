package authn_dto

type ForgotPasswordDto struct {
	Email               string `json:"email" form:"email" binding:"required,email"`
	CodeChallenge       string `json:"code_challenge,omitempty" form:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method,omitempty" form:"code_challenge_method"`
}

type ResetPasswordDto struct {
	Token        string `json:"token" form:"token" binding:"required"`
	Password     string `json:"password" form:"password" binding:"required,min=8"`
	CodeVerifier string `json:"code_verifier,omitempty" form:"code_verifier"`
}
