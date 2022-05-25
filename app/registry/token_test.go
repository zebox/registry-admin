package registry

import (
	"github.com/docker/libtrust"
	log "github.com/go-pkgz/lgr"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestNewRegistryToken(t *testing.T) {
	privateKey, errKey := libtrust.GenerateRSA2048PrivateKey()
	require.NoError(t, errKey)
	publicKey, errPubKey := libtrust.FromCryptoPublicKey(privateKey.CryptoPublicKey())
	require.NoError(t, errPubKey)

	// test with defaults
	rt, err := NewRegistryToken(
		privateKey,
		publicKey,
		"super-secret-password")
	require.NoError(t, err)
	assert.Equal(t, int64(defaultTokenExpiration), rt.tokenExpiration)
	assert.Equal(t, defaultTokenIssuer, rt.tokenIssuer)
	assert.Equal(t, privateKeyName, rt.keyName)
	assert.Equal(t, publicKeyName, rt.publicKeyName)
	assert.Equal(t, CAName, rt.CARootName)

	// test with options
	rt, err = NewRegistryToken(
		privateKey,
		publicKey,
		"super-secret-password",
		TokenExpiration(10),
		TokenIssuer("127.0.0.2"),
		TokenLogger(log.Default()),
		CertsName("test.key", "test.pub", "test_ca.crt"),
	)

	require.NoError(t, err)
	assert.Equal(t, int64(10), rt.tokenExpiration)
	assert.Equal(t, "127.0.0.2", rt.tokenIssuer)
	assert.Equal(t, "/test.key", rt.keyName)
	assert.Equal(t, "/test.pub", rt.publicKeyName)
	assert.Equal(t, "/test_ca.crt", rt.CARootName)

	// test with error
	_, err = NewRegistryToken(
		privateKey,
		publicKey,
		"abc",
		TokenExpiration(0))
	require.Error(t, err)

}

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

	claims := jwt.MapClaims{}
	testToken, errToken := jwt.ParseWithClaims(jwtToken.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return publicKey.CryptoPublicKey(), nil
	})
	assert.NoError(t, errToken)
	assert.True(t, testToken.Valid)
	assert.Equal(t, rt.tokenIssuer, claims["iss"])
	assert.Equal(t, authReq.Account, claims["sub"])
	assert.Equal(t, authReq.Service, claims["aud"])

}

func TestRegistryToken_CreateCerts(t *testing.T) {

	tmpPath := os.TempDir()

	rt := registryToken{
		keyName:       privateKeyName,
		publicKeyName: publicKeyName,
		CARootName:    CAName,
	}

	err := rt.CreateCerts(tmpPath)
	require.NoError(t, err)

	_, err = libtrust.LoadKeyFile(tmpPath + privateKeyName)
	assert.NoError(t, err)

	_, err = libtrust.LoadPublicKeyFile(tmpPath + publicKeyName)
	assert.NoError(t, err)

	_, err = libtrust.LoadCertificateBundle(tmpPath + CAName)
	assert.NoError(t, err)

	// test with error when certs exist
	err = rt.CreateCerts(tmpPath)
	assert.Error(t, err)
	assert.NoError(t, os.Remove(tmpPath+privateKeyName))

	err = rt.CreateCerts(tmpPath)
	assert.Error(t, err)
	assert.NoError(t, os.Remove(tmpPath+publicKeyName))

	err = rt.CreateCerts(tmpPath)
	assert.Error(t, err)
	assert.NoError(t, os.Remove(tmpPath+CAName))

	// test  with error when path error
	err = rt.CreateCerts(tmpPath + "unknown_path")
	assert.Error(t, err)

	rt.CARootName = ""
	err = rt.CreateCerts(tmpPath)
	assert.Error(t, err)

	rt.publicKeyName = ""
	err = rt.CreateCerts(tmpPath)
	assert.Error(t, err)

	rt.keyName = ""
	err = rt.CreateCerts(tmpPath)
	assert.Error(t, err)
}
