package types

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
	UserStatusLocked   UserStatus = "locked"
)

type KeyStatus string

const (
	KeyStatusActive   KeyStatus = "active"
	KeyStatusPrevious KeyStatus = "previous"
	KeyStatusRetired  KeyStatus = "retired"
)

type SPProtocol string

const (
	SPProtocolOIDC SPProtocol = "oidc"
	SPProtocolSAML SPProtocol = "saml"
)

type AuditResult string

const (
	AuditResultOK  AuditResult = "ok"
	AuditResultErr AuditResult = "err"
)
