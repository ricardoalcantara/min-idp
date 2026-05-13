import { SAML } from "@node-saml/node-saml"

function requireEnv(key: string): string {
  const v = process.env[key]
  if (!v) throw new Error(`Missing required env var: ${key}`)
  return v
}

// Restore real newlines if the PEM was stored as \n in the env file.
function parsePem(raw: string) {
  return raw.replace(/\\n/g, "\n")
}

export function createSAML() {
  const signRequests = process.env.SAML_SP_SIGN_REQUESTS !== "false"
  return new SAML({
    callbackUrl:             requireEnv("SAML_SP_ACS_URL"),
    entryPoint:              requireEnv("SAML_IDP_SSO_URL"),
    issuer:                  requireEnv("SAML_SP_ENTITY_ID"),
    idpCert:                 parsePem(requireEnv("SAML_IDP_CERT")),
    privateKey:              signRequests ? parsePem(requireEnv("SAML_SP_PRIVATE_KEY")) : undefined,
    publicCert:              signRequests ? parsePem(requireEnv("SAML_SP_CERT")) : undefined,
    signatureAlgorithm:      "sha256",
    wantAuthnResponseSigned: false,
    wantAssertionsSigned:    true,
    audience:                false,
  })
}

export const spEntityId = () => process.env.SAML_SP_ENTITY_ID ?? "http://localhost:3002"
export const spCert     = () => parsePem(process.env.SAML_SP_CERT ?? "")
export const idpSloUrl  = () => process.env.SAML_IDP_SLO_URL ?? ""
