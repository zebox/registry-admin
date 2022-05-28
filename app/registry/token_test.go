package registry

import (
	"github.com/docker/libtrust"
	log "github.com/go-pkgz/lgr"
	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func TestNewRegistryToken(t *testing.T) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		require.NoError(t, err)
	}
	path := filepath.ToSlash(userHomeDir) + "/" + certsDirName // fix backslashes for Windows path

	defer func() {
		assert.NoError(t, os.RemoveAll(path))
	}()

	// test with defaults with generate
	rt, err := NewRegistryToken("super-secret-password")
	require.NoError(t, err)
	assert.Equal(t, int64(defaultTokenExpiration), rt.tokenExpiration)
	assert.Equal(t, defaultTokenIssuer, rt.tokenIssuer)
	assert.Equal(t, path+privateKeyName, rt.KeyPath)
	assert.Equal(t, path+publicKeyName, rt.PublicKeyPath)
	assert.Equal(t, path+CAName, rt.CARootPath)

	// test with loading exist certs
	rt, err = NewRegistryToken("super-secret-password")
	require.NoError(t, err)

	// test with options
	tmpDir, errDir := os.MkdirTemp("", "test_cert")
	require.NoError(t, errDir)
	rt, err = NewRegistryToken(

		"super-secret-password",
		TokenExpiration(10),
		TokenIssuer("127.0.0.2"),
		TokenLogger(log.Default()),
		CertsName(Certs{
			tmpDir,
			tmpDir + "/test.key",
			tmpDir + "/test.pub",
			tmpDir + "/test_ca.crt",
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, int64(10), rt.tokenExpiration)
	assert.Equal(t, "127.0.0.2", rt.tokenIssuer)
	assert.Equal(t, tmpDir+"/test.key", rt.KeyPath)
	assert.Equal(t, tmpDir+"/test.pub", rt.PublicKeyPath)
	assert.Equal(t, tmpDir+"/test_ca.crt", rt.CARootPath)

	// test with error
	_, err = NewRegistryToken(
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
	tmpDir, errDir := os.MkdirTemp("test_", "cert")
	require.NoError(t, errDir)

	rt := registryToken{}

	rt.RootPath = tmpDir
	rt.KeyPath = tmpDir + privateKeyName
	rt.PublicKeyPath = tmpDir + publicKeyName
	rt.CARootPath = tmpDir + CAName

	err := rt.CreateCerts()
	require.NoError(t, err)

	defer func() {
		_ = os.Remove(tmpPath + CAName)
		_ = os.Remove(tmpPath + privateKeyName)
		_ = os.Remove(tmpPath + publicKeyName)
	}()

	_, err = libtrust.LoadKeyFile(tmpPath + privateKeyName)
	assert.NoError(t, err)

	_, err = libtrust.LoadPublicKeyFile(tmpPath + publicKeyName)
	assert.NoError(t, err)

	_, err = libtrust.LoadCertificateBundle(tmpPath + CAName)
	assert.NoError(t, err)

	// test with error when certs exist
	err = rt.CreateCerts()
	assert.Error(t, err)
	assert.NoError(t, os.Remove(tmpPath+privateKeyName))

	err = rt.CreateCerts()
	assert.Error(t, err)
	assert.NoError(t, os.Remove(tmpPath+publicKeyName))

	err = rt.CreateCerts()
	assert.Error(t, err)
	assert.NoError(t, os.Remove(tmpPath+CAName))

	// test  with error when path error
	err = rt.CreateCerts()
	assert.Error(t, err)

	rt.Certs.KeyPath = "*"
	err = rt.CreateCerts()
	assert.Error(t, err)

	rt.Certs.PublicKeyPath = "*"
	err = rt.CreateCerts()
	assert.Error(t, err)
	assert.NoError(t, os.Remove(tmpPath+privateKeyName))

	rt.Certs.KeyPath = privateKeyName
	rt.Certs.PublicKeyPath = publicKeyName

	rt.Certs.CARootPath = "*"
	err = rt.CreateCerts()
	assert.Error(t, err)
}
