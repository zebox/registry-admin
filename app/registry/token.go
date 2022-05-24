// Token need for authenticate and manage authorizations clients using a separate access control manager.
// A service is used by the official Docker Registry to authenticate clients and verify their authorization to Docker image repositories.
// A client should contact the registry first. If the registry server requires authentication it will return a 401 Unauthorized response with a WWW-Authenticate header
// with details how to authenticate to registry. After authenticate is successfully service will issue an opaque Bearer registryToken that clients should supply to subsequent requests
// in the Authorization header.

package registry

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"math/rand"
	"strings"
	"time"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"

	log "github.com/go-pkgz/lgr"
)

const (
	defaultTokenIssuer     = "127.0.0.1"
	defaultTokenExpiration = 60
)

// AuthorizationRequest is the authorization request data from registry when client auth call
// for detailed description go to https://docs.docker.com/registry/spec/auth/jwt/
type AuthorizationRequest struct {

	// Bind to 'sub' token header
	// The subject of the token; the name or id of the client which requested it.
	// This should be empty (`""`) if the client did not authenticate.
	Account string

	// Bind to token 'aud' header. The intended audience of the token; the name or id of the service which will verify
	// the token to authorize the client/subject.
	Service string

	// The subject of the token; the name or id of the client which requested it.
	// This should be empty (`""`) if the client did not authenticate.
	Type string

	// The name of the resource of the given type hosted by the service.
	Name string

	// An array of strings which give the actions authorized on this resource.
	Actions []string

	IP string
}

// clientToken is Bearer registryToken representing authorized access for a client
type clientToken struct {
	// An opaque Bearer registryToken that clients should supply to subsequent requests in the Authorization header.
	Token string `json:"registryToken"`

	// For compatibility with OAuth 2.0, we will also accept registryToken under the name access_token.
	// At least one of these fields must be specified, but both may also appear (for compatibility with older clients).
	// When both are specified, they should be equivalent; if they differ the client's choice is undefined.
	AccessToken string `json:"access_token"`
}

type registryToken struct {

	// registryToken claims field
	tokenIssuer string

	// registryToken life
	tokenExpiration int64

	// main secret phrase for registryToken signing
	secret string

	// keys pair for generate JWT signature
	privateKey libtrust.PrivateKey
	publicKey  libtrust.PublicKey

	l log.L
}
type TokenOption func(option *registryToken)

// TokenExpiration option define custom token expiration time
func TokenExpiration(expirationTime int64) TokenOption {
	return func(rt *registryToken) {
		rt.tokenExpiration = expirationTime
	}
}

// TokenIssuer option define token issuer, typically the fqdn of the authorization server
func TokenIssuer(issuer string) TokenOption {
	return func(rt *registryToken) {
		rt.tokenIssuer = issuer
	}
}

// TokenLogger define logger instance
func TokenLogger(l log.L) TokenOption {
	return func(rt *registryToken) {
		rt.l = l
	}
}

// NewRegistryToken will construct new tokenRegistry instance with required options
// and allow re-define default option for token generator
func NewRegistryToken(key libtrust.PrivateKey, publicKey libtrust.PublicKey, secretPhrase string, opts ...TokenOption) (*registryToken, error) {
	rt := &registryToken{
		secret:          secretPhrase,
		privateKey:      key,
		publicKey:       publicKey,
		tokenExpiration: defaultTokenExpiration,
		tokenIssuer:     defaultTokenIssuer,
		l:               log.Default(),
	}
	for _, opt := range opts {
		opt(rt)
	}

	if len(secretPhrase) < 10 {
		log.Print("[WARN] the secret for token sign is weak\n")
	}

	if rt.tokenExpiration < 1 {
		return nil, errors.Errorf("token expiration time is invalid, should great more than one")
	}
	return rt, nil
}

func (rt *registryToken) Generate(authRequest *AuthorizationRequest) (clientToken, error) {
	// sign any string to get the used signing Algorithm for the private key
	_, algo, err := rt.privateKey.Sign(strings.NewReader(rt.secret), 0)

	if err != nil {
		return clientToken{}, err
	}

	header := token.Header{
		Type:       "JWT",
		SigningAlg: algo,
		KeyID:      rt.publicKey.KeyID(),
	}

	headerJson, err := json.Marshal(header)
	if err != nil {
		return clientToken{}, err
	}
	now := time.Now().Unix()
	expr := now + defaultTokenExpiration

	if rt.tokenExpiration > 0 {
		expr = now + rt.tokenExpiration
	}

	// init default registryToken claims
	claim := token.ClaimSet{
		Issuer:     rt.tokenIssuer,
		Subject:    authRequest.Account,
		Audience:   authRequest.Service,
		Expiration: expr,
		NotBefore:  now - 10,
		IssuedAt:   now,
		JWTID:      fmt.Sprintf("%d", rand.Intn(4096-1024)+1024),
		Access:     []*token.ResourceActions{},
	}

	claim.Access = append(claim.Access, &token.ResourceActions{
		Type:    authRequest.Type,
		Name:    authRequest.Name,
		Actions: authRequest.Actions,
	})

	claimJson, err := json.Marshal(claim)
	if err != nil {
		return clientToken{}, err
	}

	encodeToBase64 := func(b []byte) string {
		return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
	}

	payload := fmt.Sprintf("%s%s%s", encodeToBase64(headerJson), token.TokenSeparator, encodeToBase64(claimJson))
	sig, sigAlgo, err := rt.privateKey.Sign(strings.NewReader(payload), 0)
	if err != nil && sigAlgo != algo {
		return clientToken{}, err
	}
	tokenString := fmt.Sprintf("%s%s%s", payload, token.TokenSeparator, encodeToBase64(sig))
	return clientToken{Token: tokenString, AccessToken: tokenString}, nil
}
