package registry

// Token need for authenticate and manage authorizations clients using a separate access control manager.
// A service is used by the official Docker Registry to authenticate clients and verify their authorization to Docker image repositories.
// A client should contact the registry first. If the registry server requires authentication it will return a 401 Unauthorized response with a WWW-Authenticate header
// with details how to authenticate to registry. After authenticate is successfully service will issue an opaque Bearer registryToken that clients should supply to subsequent requests
// in the Authorization header. More details by link https://docs.docker.com/registry/spec/auth/jwt/

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"

	log "github.com/go-pkgz/lgr"
)

const (
	defaultTokenIssuer     = "127.0.0.1"
	defaultTokenExpiration = 60

	// default names of generated certificate
	certsDirName   = "./registry-certs"
	privateKeyName = "/registry_auth.key"
	publicKeyName  = "/registry_auth.pub"
	CAName         = "/registry_auth_ca.crt"
)

var ErrTemplateCertFileAlreadyExist = "cert file '%s' already exist"

// Certs will define a path to certs either for loading private, public and CARoot files or path to save ones when CreateCerts call.
// CreateCerts doesn't overwrite existed files in a path, user should delete them before method call.
type Certs struct {
	RootPath      string
	KeyPath       string
	PublicKeyPath string
	CARootPath    string
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
	Certs

	// registryToken claims field
	tokenIssuer string

	// registryToken life
	tokenExpiration int64

	// main secret phrase for registryToken signing
	secret string

	// certificates for generate JWT signature
	privateKey libtrust.PrivateKey
	publicKey  libtrust.PublicKey
	caRoot     *x509.Certificate

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

// CertsName define custom certs file name
func CertsName(certs Certs) TokenOption {
	return func(rt *registryToken) {
		rt.RootPath = certs.CARootPath
		rt.PublicKeyPath = certs.PublicKeyPath
		rt.KeyPath = certs.KeyPath
		rt.CARootPath = certs.CARootPath
	}
}

// NewRegistryToken will construct new tokenRegistry instance with required options
// and allow re-define default option for token generator
func NewRegistryToken(secretPhrase string, opts ...TokenOption) (*registryToken, error) {

	rt := &registryToken{
		secret:          secretPhrase,
		tokenExpiration: defaultTokenExpiration,
		tokenIssuer:     defaultTokenIssuer,
		l:               log.Default(),
	}

	// define default certificate files path
	rt.RootPath = certsDirName
	rt.PublicKeyPath = certsDirName + privateKeyName
	rt.KeyPath = certsDirName + publicKeyName
	rt.CARootPath = certsDirName + CAName

	for _, opt := range opts {
		opt(rt)
	}

	if len(secretPhrase) < 10 {
		log.Print("[WARN] the secret for token sign is weak\n")
	}

	if rt.tokenExpiration < 1 {
		return nil, errors.Errorf("token expiration time is invalid, should great more than one")
	}

	if err := rt.loadCerts(); err != nil {
		err = rt.CreateCerts()
		if err != nil {
			return nil, err
		}
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
		JWTID:      fmt.Sprintf("%d", rand.Intn(4096-1024)+1024), // nolint
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

func (rt *registryToken) CreateCerts() (err error) {

	rt.privateKey, err = libtrust.GenerateRSA2048PrivateKey()
	if err != nil {
		return err
	}

	rt.publicKey, err = libtrust.FromCryptoPublicKey(rt.privateKey.CryptoPublicKey())
	if err != nil {
		return err
	}

	rt.caRoot, err = libtrust.GenerateCACert(rt.privateKey, rt.publicKey)
	if err != nil {
		return err
	}

	return rt.saveKeys()
}

func (rt *registryToken) loadCerts() (err error) {

	if _, err = os.Stat(rt.KeyPath); err != nil {
		return err
	}

	rt.privateKey, err = libtrust.LoadKeyFile(rt.KeyPath)
	if err != nil {
		return err
	}

	rt.publicKey, err = libtrust.LoadKeyFile(rt.PublicKeyPath)
	if err != nil {
		return err
	}

	bundle, errCaLoad := libtrust.LoadCertificateBundle(rt.CARootPath)
	if errCaLoad != nil {
		return errCaLoad
	}
	rt.caRoot = bundle[0]

	return nil
}

func (rt registryToken) saveKeys() error {

	var errExist error
	// check if certs already exist
	if _, err := os.Stat(rt.KeyPath); err == nil {
		err = errors.Errorf(ErrTemplateCertFileAlreadyExist, rt.KeyPath)
		errExist = multierror.Append(errExist, err)
	}

	if _, err := os.Stat(rt.PublicKeyPath); err == nil {
		err = errors.Errorf(ErrTemplateCertFileAlreadyExist, rt.PublicKeyPath)
		errExist = multierror.Append(errExist, err)
	}

	if _, err := os.Stat(rt.CARootPath); err == nil {
		err = errors.Errorf(ErrTemplateCertFileAlreadyExist, rt.CARootPath)
		errExist = multierror.Append(errExist, err)

	}

	if errExist != nil {
		return errExist
	}

	// trying save new certs to files
	if err := libtrust.SaveKey(rt.KeyPath, rt.privateKey); err != nil {
		return errors.Wrap(err, "failed to parse private key to PEM")
	}

	if err := libtrust.SavePublicKey(rt.PublicKeyPath, rt.publicKey); err != nil {
		return errors.Wrap(err, "failed to save public key to file")
	}

	err := ioutil.WriteFile(rt.CARootPath, rt.caRoot.Raw, 0644) // nolint:gosec
	if err != nil {
		return errors.Wrap(err, "failed to save CA bundle to file")
	}
	return nil
}
