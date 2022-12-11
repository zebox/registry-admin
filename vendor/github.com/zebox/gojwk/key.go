package gojwk

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"github.com/pkg/errors"
)

type keyStorage interface {
	Load() (*rsa.PrivateKey, error)                // implement loader for a private RSA Keys pairs from storage provider
	Save(key *rsa.PrivateKey, certCA []byte) error // implement save a Keys pairs and root certificates bundle to storage provider
}

// Keys using for create and validate token signature
type Keys struct {

	// Keys identification for detect and use Keys
	KeyID string

	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey

	//  Certificate and  Certificate Authority
	certCARoot []byte

	// Keys bit size value, set in Options, default - 2048
	bitSize int

	// define saver and loader Keys pair function
	// storage required has path to public and private Keys file which will load from disk
	storage keyStorage
}

// NewKeys create new Keys pair
func NewKeys(options ...Options) (keysPair *Keys, err error) {

	// define Keys and default values
	keysPair = &Keys{
		bitSize: 2048,
	}

	// parse keysPair options
	for _, opt := range options {
		opt(keysPair)
	}

	// force encrypt with Keys 128-bits or more
	if keysPair.bitSize < 128 {
		return nil, errors.New("bit size invalid and should has length 128 or more")
	}

	return keysPair, nil
}

// Generate new keys pair and save if external storage field defined
func (k *Keys) Generate() (err error) {
	reader := rand.Reader

	if k.privateKey, err = rsa.GenerateKey(reader, k.bitSize); err != nil {
		return errors.Wrapf(err, "failed to generate new Keys pair")
	}

	k.publicKey = &k.privateKey.PublicKey
	k.KeyID = k.kid()

	return nil
}

// Save keys pair to provider storage if it defined
func (k *Keys) Save() error {
	// check for external Keys storage defined and try save new Keys
	if k.storage != nil {
		return k.storage.Save(k.privateKey, k.certCARoot)
	}
	return errors.New("storage provider undefined")
}

// Load trying loading private and public key pair from storage provider
func (k *Keys) Load() (err error) {

	if k.storage == nil {
		return errors.New("failed to load key pair, storage provider undefined")
	}

	if k.privateKey, err = k.storage.Load(); err != nil {
		return errors.Wrap(err, "failed to load private key")
	}
	// assign public key from private
	k.publicKey = &k.privateKey.PublicKey
	return nil
}

// JWK create JSON Web Keys from public Keys
func (k *Keys) JWK() (jwk JWK, err error) {
	return NewJWK(k.publicKey)
}

// Private return private Keys for sign jwt
func (k *Keys) Private() *rsa.PrivateKey {
	return k.privateKey
}

func (k *Keys) CreateCAROOT(ca *x509.Certificate) error {
	if k.privateKey == nil || k.publicKey == nil {
		return errors.New("private and public keys shouldn't be nil when CA create")
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, k.publicKey, k.privateKey)
	if err != nil {
		return errors.Wrap(err, "failed to create certificate")
	}

	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	if err != nil {
		return errors.Wrap(err, "failed to encode certificate CA to PEM bytes")
	}

	caPrivKeyPEM := new(bytes.Buffer)
	err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(k.privateKey),
	})
	if err != nil {
		return errors.Wrap(err, "failed to encode private key certificate to PEM bytes")
	}

	// assign certificates bytes to CA ROOT field
	k.certCARoot = caPEM.Bytes()

	return nil
}

// CertCA return CA root certificate
func (k *Keys) CertCA() []byte {
	return k.certCARoot
}

func PEMBytes(key interface{}) ([]byte, error) {
	switch key.(type) {
	case *rsa.PrivateKey:
		keyBytes, err := x509.MarshalPKCS8PrivateKey(key)
		if err != nil {
			return nil, errors.New("failed to marshal private key to bytes")
		}
		return pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: keyBytes,
			},
		), nil
	case *rsa.PublicKey:
		keyBytes, err := x509.MarshalPKIXPublicKey(key)
		if err != nil {
			return nil, errors.Wrap(err, "failed parse public key to PEM bytes encode")
		}
		return pem.EncodeToMemory(
			&pem.Block{
				Type:  "RSA PUBLIC KEY",
				Bytes: keyBytes,
			},
		), nil
	}
	return nil, errors.New("failed key file type for bytes encode")
}

// kid return Keys ID of public key for map with JWK
func (k *Keys) kid() string {
	n := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(k.publicKey.N.Bytes())

	// create kid from public Keys modulus
	h := sha256.New()
	h.Write([]byte(n))
	kidBytes := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(kidBytes)[:4]

}
