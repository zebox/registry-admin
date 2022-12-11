// The JSON Web Keys Set (JWKS) is a set of keys containing the public keys that should be used to verify any JSON Web Token (JWT)
// that is issued by an authorization server and signed using the RSA or ECDSA algorithms.
// This package implement tool for generate, loading and save keys pair for use as JWK
// For more information see https://datatracker.ietf.org/doc/html/rfc7517

package gojwk

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"math/big"
)

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS is a list of JWK keys
type JWKS []JWK

// NewJWK is main constructor for create JWK from raw public Keys, accept pointer to *rsa.PublicKey
func NewJWK(publicKey *rsa.PublicKey) (jwk JWK, err error) {
	if publicKey == nil {
		return jwk, errors.New("public Keys should be defined")
	}

	// convert to modulus
	n := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(publicKey.N.Bytes())

	// convert to exponent
	eBuff := make([]byte, 4)
	binary.LittleEndian.PutUint32(eBuff, uint32(publicKey.E))
	e := base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(eBuff)

	// create kid from public Keys modulus
	h := sha256.New()
	h.Write([]byte(n))
	kidBytes := h.Sum(nil)
	kid := base64.StdEncoding.EncodeToString(kidBytes)

	jwk = JWK{Alg: "RS256", Kty: "RSA", Use: "sig", Kid: kid[:4], N: n, E: e[:4]}

	return jwk, nil
}

// PublicKey return raw public Keys from JWK
func (j *JWK) PublicKey() (*rsa.PublicKey, error) {
	return j.parsePublicKey()
}

// parsePublicKey from JWK to RSA public Keys
func (j *JWK) parsePublicKey() (*rsa.PublicKey, error) {

	bufferN, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(j.N) // decode modulus
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode public Keys modulus (n)")
	}

	bufferE, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(j.E) // decode exponent
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode public Keys exponent (e)")
	}

	// create rsa public Keys from JWK data
	publicKey := &rsa.PublicKey{
		N: big.NewInt(0).SetBytes(bufferN),
		E: int(big.NewInt(0).SetBytes(bufferE).Int64()),
	}
	return publicKey, nil
}

// ToString convert JWK object to JSON string
func (j *JWK) ToString() string {
	jwkBuffer, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return ""
	}
	return string(jwkBuffer)
}

// ToString convert JWKS item list JSON string
func (js JWKS) ToString() string {

	jwksBuffer, err := json.Marshal(js)
	if err != nil {
		return ""
	}
	return string(jwksBuffer)
}

// KeyFunc use for JWT sign verify with specific public Keys
func (j *JWK) KeyFunc(token *jwt.Token) (interface{}, error) {

	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("get JWT kid header not found")
	}
	if j.Kid != keyID {
		return nil, errors.Errorf("hasn't JWK with kid [%s] for check", keyID)
	}
	publicKey, err := j.parsePublicKey()
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}
