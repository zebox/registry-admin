package registry

// Token need for authenticate and manage authorizations clients using a separate access control manager.
// A service is used by the official Docker Registry to authenticate clients and verify their authorization to Docker image repositories.
// A client should contact the registry first. If the registry server requires authentication it will return a 401 Unauthorized response with a WWW-Authenticate header
// with details how to authenticate to registry. After authenticate is successfully service will issue an opaque Bearer AccessToken that clients should supply to subsequent requests
// in the Authorization header. More details by link https://docs.docker.com/registry/spec/auth/jwt/

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	crand "crypto/rand"
	log "github.com/go-pkgz/lgr"
)

const (
	defaultTokenIssuer     = "127.0.0.1"
	defaultTokenExpiration = 60

	// default names of generated certificate
	certsDirName   = ".registry-certs"
	privateKeyName = "/registry_auth.key"
	publicKeyName  = "/registry_auth.pub"
	caName         = "/registry_auth_ca.crt"

	errPrefixCertsNotFound = "cert file not found"
)

var errTemplateCertFileAlreadyExist = "cert file '%s' already exist"

// TokenRequest is the authorization request data from registry when client auth call
// for detailed description go to https://docs.docker.com/registry/spec/auth/jwt/
type TokenRequest struct {

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

	// Custom TTL for a new token
	ExpireTime int64
}

// Certs will define a path to certs either for loading private, public and CARoot files or path to save ones when createCerts call.
// createCerts doesn't overwrite existed files in a path, user should delete them before method call.
type Certs struct {
	RootPath      string
	KeyPath       string
	PublicKeyPath string
	CARootPath    string
	FQDNs         []string
	IP            string
}

// ClientToken is Bearer AccessToken representing authorized access for a client
type ClientToken struct {
	// An opaque Bearer token that clients should supply to subsequent requests in the Authorization header.
	Token string `json:"token"`

	// For compatibility with OAuth 2.0, we will also accept AccessToken under the name access_token.
	// At least one of these fields must be specified, but both may also appear (for compatibility with older clients).
	// When both are specified, they should be equivalent; if they differ the client's choice is undefined.
	AccessToken string `json:"access_token"`
}

// AccessToken is a token instance using for authorization in registryal
type AccessToken struct {
	Certs

	// AccessToken claims field
	tokenIssuer string

	// AccessToken life
	tokenExpiration int64

	// certificates for generate JWT signature
	privateKey libtrust.PrivateKey
	publicKey  libtrust.PublicKey
	caRoot     *x509.Certificate

	l log.L
}

// TokenOption defines options which pass to token
type TokenOption func(option *AccessToken)

// TokenExpiration option define custom token expiration time
func TokenExpiration(expirationTime int64) TokenOption {
	return func(rt *AccessToken) {
		rt.tokenExpiration = expirationTime
	}
}

// TokenIssuer option define token issuer, typically the fqdn of the authorization server
func TokenIssuer(issuer string) TokenOption {
	if issuer == "" {
		issuer = defaultTokenIssuer
	}
	return func(rt *AccessToken) {
		rt.tokenIssuer = issuer
	}
}

// TokenLogger define logger instance
func TokenLogger(l log.L) TokenOption {
	return func(rt *AccessToken) {
		rt.l = l
	}
}

// CertsName define custom certs file name
func CertsName(certs Certs) TokenOption {
	return func(rt *AccessToken) {
		rt.RootPath = certs.RootPath
		rt.PublicKeyPath = certs.PublicKeyPath
		rt.KeyPath = certs.KeyPath
		rt.CARootPath = certs.CARootPath
		rt.FQDNs = certs.FQDNs
		rt.IP = certs.IP
	}
}

// NewRegistryToken will construct new tokenRegistry instance with required options
// and allow re-define default option for token generator
func NewRegistryToken(opts ...TokenOption) (*AccessToken, error) {

	rt := &AccessToken{
		tokenExpiration: defaultTokenExpiration,
		tokenIssuer:     defaultTokenIssuer,
		l:               log.Default(),
	}

	// Create default directory where certificates will be created by default.
	// The directory default path is a home directory at user which process executed app
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to obtain home directory for user which process run")
	}
	path := filepath.ToSlash(userHomeDir) + "/" + certsDirName // fix backslashes for Windows path

	// define certificate files path if their options omitted
	rt.RootPath = path
	rt.PublicKeyPath = rt.RootPath + publicKeyName
	rt.KeyPath = rt.RootPath + privateKeyName
	rt.CARootPath = rt.RootPath + caName
	rt.FQDNs = []string{"localhost"}

	for _, opt := range opts {
		opt(rt)
	}

	if err = os.Mkdir(path, os.ModePerm); err != nil && !os.IsExist(err) {
		return nil, errors.Wrap(err, "failed to create default directory for save certificates")
	}

	if rt.tokenExpiration < 1 {
		return nil, errors.Errorf("token expiration time is invalid, should great more than one")
	}

	if err = rt.loadCerts(); err != nil {

		// throw error when certs files exist but load is fail, otherwise tries to generate new certs
		if !strings.HasPrefix(err.Error(), errPrefixCertsNotFound) {
			return nil, err
		}

		err = rt.createCerts()
		if err != nil {
			return nil, err
		}
	}

	return rt, nil
}

func (rt *AccessToken) generate(tokenRequest *TokenRequest) (ClientToken, error) {
	// sign any string to get the used signing Algorithm for the private key
	_, algo, err := rt.privateKey.Sign(strings.NewReader(""), 0)

	if err != nil {
		return ClientToken{}, err
	}

	header := token.Header{
		Type:       "JWT",
		SigningAlg: algo,
		KeyID:      rt.publicKey.KeyID(),
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return ClientToken{}, err
	}
	now := time.Now().Unix()
	expr := now + defaultTokenExpiration

	if rt.tokenExpiration > 0 {
		expr = now + rt.tokenExpiration
	}

	// custom token expiration time should more or equal 60 seconds.
	if tokenRequest.ExpireTime >= defaultTokenExpiration {
		expr = now + tokenRequest.ExpireTime
	}

	// init default AccessToken claims
	claim := token.ClaimSet{
		Issuer:     rt.tokenIssuer,
		Subject:    tokenRequest.Account,
		Audience:   tokenRequest.Service,
		Expiration: expr,
		NotBefore:  now - 10,
		IssuedAt:   now,
		JWTID:      fmt.Sprintf("%d", rand.Intn(4096-1024)+1024), // nolint
		Access:     []*token.ResourceActions{},
	}

	claim.Access = append(claim.Access, &token.ResourceActions{
		Type:    tokenRequest.Type,
		Name:    tokenRequest.Name,
		Actions: tokenRequest.Actions,
	})

	claimJSON, err := json.Marshal(claim)
	if err != nil {
		return ClientToken{}, err
	}

	encodeToBase64 := func(b []byte) string {
		return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
	}

	payload := fmt.Sprintf("%s%s%s", encodeToBase64(headerJSON), token.TokenSeparator, encodeToBase64(claimJSON))
	sig, sigAlgo, err := rt.privateKey.Sign(strings.NewReader(payload), 0)
	if err != nil && sigAlgo != algo {
		return ClientToken{}, err
	}

	tokenString := fmt.Sprintf("%s%s%s", payload, token.TokenSeparator, encodeToBase64(sig))
	return ClientToken{Token: tokenString, AccessToken: tokenString}, nil
}

func (rt *AccessToken) createCerts() (err error) {

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

	rt.appendDSnToCertificate()

	return rt.saveKeys()
}

// statCerts - checks files defined in certs option for exist
func (rt *AccessToken) statCerts() error {
	var (
		errExist  error
		errString []string
	)

	if _, err := os.Stat(rt.KeyPath); err != nil {
		errString = append(errString, rt.KeyPath)
	}

	if _, err := os.Stat(rt.PublicKeyPath); err != nil {
		errString = append(errString, rt.PublicKeyPath)
	}

	if _, err := os.Stat(rt.CARootPath); err != nil {
		errString = append(errString, rt.CARootPath)
	}

	if len(errString) > 0 {
		errExist = fmt.Errorf("%s: %s", errPrefixCertsNotFound, strings.Join(errString, ", "))
	}

	return errExist
}

func (rt *AccessToken) loadCerts() (err error) {

	if errStat := rt.statCerts(); errStat != nil {
		return errStat
	}

	rt.privateKey, err = libtrust.LoadKeyFile(rt.Certs.KeyPath)
	if err != nil {
		return err
	}

	rt.publicKey, err = libtrust.LoadPublicKeyFile(rt.Certs.PublicKeyPath)
	if err != nil {
		return err
	}

	bundle, errCaLoad := libtrust.LoadCertificateBundle(rt.Certs.CARootPath)
	if errCaLoad != nil {

		return errCaLoad
	}

	if len(bundle) == 0 {
		return errors.New("certificates bundle not found in CA file")
	}
	rt.caRoot = bundle[0]

	return nil
}

func (rt *AccessToken) saveKeys() error {

	var errExist error
	// check if certs already exist
	if _, err := os.Stat(rt.KeyPath); err == nil {
		err = errors.Errorf(errTemplateCertFileAlreadyExist, rt.KeyPath)
		errExist = multierror.Append(errExist, err)
	}

	if _, err := os.Stat(rt.PublicKeyPath); err == nil {
		err = errors.Errorf(errTemplateCertFileAlreadyExist, rt.PublicKeyPath)
		errExist = multierror.Append(errExist, err)
	}

	if _, err := os.Stat(rt.CARootPath); err == nil {
		err = errors.Errorf(errTemplateCertFileAlreadyExist, rt.CARootPath)
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

	// generated keys hasn't Subject Alternative Name for requested IP and Domain when creating with libtrust
	// and that should add this values after certificate created and extracting raw bytes after that
	caBytes, err := x509.CreateCertificate(crand.Reader, rt.caRoot, rt.caRoot, rt.publicKey.CryptoPublicKey(), rt.privateKey.CryptoPrivateKey())
	if err != nil {
		return errors.Wrap(err, "failed to create certificate")
	}

	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	if err != nil {
		return errors.Wrap(err, "failed to encode certificate for file save")
	}

	err = os.WriteFile(rt.CARootPath, certPEM.Bytes(), os.ModePerm)
	if err != nil {
		return errors.Wrap(err, "failed to save CA bundle to file")
	}
	return nil
}

// parseToken convert token string set to ClientToken struct
func (rt *AccessToken) parseToken(tokenString string) (ct ClientToken, err error) {
	if err := json.Unmarshal([]byte(tokenString), &ct); err != nil {
		return ClientToken{}, err
	}
	return ct, nil
}

// appendDSnToCertificate appends Subject Alternative Name for requested IP and Domain to certificate
// It's require for append Subject Alternative Name for requested IP and Domain to certificate
// and prevents untrusted error with HTTPS client request
// https://oidref.com/2.5.29.17
func (rt *AccessToken) appendDSnToCertificate() {
	if rt.Certs.IP != "" {
		var ipAddressRegExp = regexp.MustCompile(`(?m)^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
		if ipAddressRegExp.MatchString(rt.Certs.IP) {
			rt.caRoot.IPAddresses = append(rt.caRoot.IPAddresses, net.ParseIP(rt.Certs.IP))
		} else {
			rt.l.Logf("failed to append ip address to certificate SN, ip address is invalid")
		}
	}

	rt.caRoot.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign
	rt.caRoot.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}

	if len(rt.FQDNs) > 0 {
		rt.caRoot.DNSNames = append(rt.caRoot.DNSNames, rt.FQDNs...)
	}
}
