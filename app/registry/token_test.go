package registry

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/docker/libtrust"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zebox/gojwk"
	"github.com/zebox/gojwk/storage"
)

const (
	testPrivateKeyFileName = "test_private.key"
	testPublicKeyFileName  = "test_public.key"
)

func TestRegistryToken_Generate(t *testing.T) {
	privateKey, errKey := libtrust.GenerateRSA2048PrivateKey()
	require.NoError(t, errKey)
	publicKey, errPubKey := libtrust.FromCryptoPublicKey(privateKey.CryptoPublicKey())
	require.NoError(t, errPubKey)

	rt := registryToken{
		tokenIssuer:     "OLYMP TESTER",
		tokenExpiration: 3,
		secret:          "super-test-secret",
		privateKey:      privateKey,
		publicKey:       publicKey,
	}
	authReq := AuthorizationRequest{
		Account: "Martian",
		Service: "127.0.0.1",
		Type:    "registry",
		Name:    "test-resource",
		Actions: []string{"pull", "push"},
	}

	jwtToken, err := rt.Generate(&authReq)
	assert.NoError(t, err)

	var claims jwt.MapClaims
	testToken, errToken := jwt.ParseWithClaims(jwtToken.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return publicKey.CryptoPublicKey(), nil
	})
	assert.NoError(t, errToken)

	iat := testToken.Header["iat"].(int64)
	exp := testToken.Header["exp"].(int64)
	assert.Equal(t, exp, iat+rt.tokenExpiration)
}

func initTestKeys(ctx context.Context, t *testing.T) (keys *gojwk.Keys, dir string, err error) {

	dir, err = ioutil.TempDir(os.TempDir(), "tk")
	if err != nil {
		return nil, "", err
	}

	fileStore := storage.NewFileStorage(dir, testPrivateKeyFileName, testPublicKeyFileName)
	keys, _ = gojwk.NewKeys(gojwk.Storage(fileStore))

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{

			Organization:  []string{"OLYMP, INC."},
			Country:       []string{"SPC"},
			Province:      []string{""},
			Locality:      []string{"Mars"},
			StreetAddress: []string{"Mariner valley"},
			PostalCode:    []string{"000001"},
		},

		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// add Subject Alternative Name for requested IP and Domain
	// it prevents untasted error with client request
	// https://oidref.com/2.5.29.17
	ca.IPAddresses = append(ca.IPAddresses, net.ParseIP("127.0.0.1"))
	ca.IPAddresses = append(ca.IPAddresses, net.ParseIP("::"))
	ca.DNSNames = append(ca.DNSNames, "localhost")

	// check keys for exist in the storage provider path
	if err = keys.Load(); err != nil {

		// if keys doesn't exist or load fail then create new
		if err = keys.Generate(); err != nil {
			return nil, "", err
		}

		// create CA certificate for created keys pair
		if err = keys.CreateCAROOT(ca); err != nil {
			return nil, "", err
		}

		// if new keys pair created successfully save they to defined storage
		if err = keys.Save(); err != nil {
			return nil, "", err
		}

	}

	if err = keys.CreateCAROOT(ca); err != nil {
		return nil, "", err
	}

	go func() {
		<-ctx.Done()
		assert.NoError(t, os.RemoveAll(dir))
	}()
	return keys, dir, nil
}
