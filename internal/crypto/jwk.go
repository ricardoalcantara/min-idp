package crypto

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/base64"
	"math/big"
)

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
	Crv string `json:"crv,omitempty"`
	X   string `json:"x,omitempty"`
	Y   string `json:"y,omitempty"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func RSAPublicKeyToJWK(pub *rsa.PublicKey, kid, alg string) JWK {
	return JWK{
		Kty: "RSA", Use: "sig", Kid: kid, Alg: alg,
		N: base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E: base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}

func ECPublicKeyToJWK(pub *ecdsa.PublicKey, kid, alg string) JWK {
	byteLen := (pub.Curve.Params().BitSize + 7) / 8
	return JWK{
		Kty: "EC", Use: "sig", Kid: kid, Alg: alg,
		Crv: pub.Curve.Params().Name,
		X:   base64.RawURLEncoding.EncodeToString(padLeft(pub.X.Bytes(), byteLen)),
		Y:   base64.RawURLEncoding.EncodeToString(padLeft(pub.Y.Bytes(), byteLen)),
	}
}

func padLeft(b []byte, size int) []byte {
	if len(b) >= size {
		return b
	}
	padded := make([]byte, size)
	copy(padded[size-len(b):], b)
	return padded
}
