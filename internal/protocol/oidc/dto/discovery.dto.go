package oidc_dto

type DiscoveryDocument struct {
	Issuer                        string   `json:"issuer"`
	AuthorizationEndpoint         string   `json:"authorization_endpoint"`
	TokenEndpoint                 string   `json:"token_endpoint"`
	UserinfoEndpoint              string   `json:"userinfo_endpoint"`
	JWKSURI                       string   `json:"jwks_uri"`
	RevocationEndpoint            string   `json:"revocation_endpoint"`
	IntrospectionEndpoint         string   `json:"introspection_endpoint"`
	EndSessionEndpoint            string   `json:"end_session_endpoint"`
	ResponseTypesSupported        []string `json:"response_types_supported"`
	SubjectTypesSupported         []string `json:"subject_types_supported"`
	IDTokenSigningAlgValues       []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported               []string `json:"scopes_supported"`
	TokenEndpointAuthMethods      []string `json:"token_endpoint_auth_methods_supported"`
	GrantTypesSupported           []string `json:"grant_types_supported"`
	ClaimsSupported               []string `json:"claims_supported"`
	CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
}

func Build(issuer string) DiscoveryDocument {
	return DiscoveryDocument{
		Issuer:                        issuer,
		AuthorizationEndpoint:         issuer + "/oauth2/authorize",
		TokenEndpoint:                 issuer + "/oauth2/token",
		UserinfoEndpoint:              issuer + "/oauth2/userinfo",
		JWKSURI:                       issuer + "/.well-known/jwks.json",
		RevocationEndpoint:            issuer + "/oauth2/revoke",
		IntrospectionEndpoint:         issuer + "/oauth2/introspect",
		EndSessionEndpoint:            issuer + "/oauth2/logout",
		ResponseTypesSupported:        []string{"code"},
		SubjectTypesSupported:         []string{"public"},
		IDTokenSigningAlgValues:       []string{"ES256", "RS256"},
		ScopesSupported:               []string{"openid", "profile", "email"},
		TokenEndpointAuthMethods:      []string{"client_secret_basic", "client_secret_post"},
		GrantTypesSupported:           []string{"authorization_code", "refresh_token"},
		ClaimsSupported:               []string{"sub", "email", "name"},
		CodeChallengeMethodsSupported: []string{"S256"},
	}
}
