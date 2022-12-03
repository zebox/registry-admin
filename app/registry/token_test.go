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
	_, err = NewRegistryToken("super-secret-password")
	require.NoError(t, err)

	// test with options
	tmpDir, errDir := os.MkdirTemp(os.TempDir(), "test_cert")
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
		ServiceIpHost("127.0.0.2", "domain.local.test"),
	)

	require.NoError(t, err)
	assert.Equal(t, int64(10), rt.tokenExpiration)
	assert.Equal(t, "127.0.0.2", rt.tokenIssuer)
	assert.Equal(t, tmpDir+"/test.key", rt.KeyPath)
	assert.Equal(t, tmpDir+"/test.pub", rt.PublicKeyPath)
	assert.Equal(t, tmpDir+"/test_ca.crt", rt.CARootPath)
	assert.Equal(t, "127.0.0.2", rt.serviceIP)
	assert.Equal(t, "domain.local.test", rt.serviceHost)

	// test with error
	_, err = NewRegistryToken(
		"abc",
		TokenExpiration(0))
	require.Error(t, err)

	// test with bad certs path
	badPath := "badPath"
	_, err = NewRegistryToken(
		"super-secret-password",
		CertsName(Certs{
			RootPath: badPath,
		}),
	)
	assert.Error(t, err)

	_, err = NewRegistryToken(
		"super-secret-password",
		CertsName(Certs{
			RootPath: tmpDir,
			KeyPath:  tmpDir + "/bad_key_path.key",
		}),
	)
	assert.Error(t, err)

	_, err = NewRegistryToken(
		"super-secret-password",
		CertsName(Certs{
			RootPath:      tmpDir,
			KeyPath:       tmpDir + "/test.key",
			PublicKeyPath: tmpDir + "/bad_public_path.pub",
		}),
	)
	assert.Error(t, err)

	_, err = NewRegistryToken(
		"super-secret-password",
		CertsName(Certs{
			RootPath:      tmpDir,
			KeyPath:       tmpDir + "/test.key",
			PublicKeyPath: tmpDir + "/test.pub",
			CARootPath:    tmpDir + "bac_ca_file_path.crt",
		}),
	)
	assert.Error(t, err)
}

func TestRegistryToken_Generate(t *testing.T) {
	tmpDir, errDir := os.MkdirTemp(os.TempDir(), "test_cert")
	require.NoError(t, errDir)
	defer func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}()

	rt, err := NewRegistryToken(

		"super-secret-password",
		TokenIssuer("OLYMP TESTER"),
		TokenExpiration(3),
		CertsName(Certs{
			tmpDir,
			tmpDir + "/test.key",
			tmpDir + "/test.pub",
			tmpDir + "/test_ca.crt",
		}),
	)
	require.NoError(t, err)

	authReq := TokenRequest{
		Account:    "Martian",
		Service:    "127.0.0.1",
		Type:       "registry",
		Name:       "test-resource",
		Actions:    []string{"pull", "push"},
		ExpireTime: 120,
	}

	jwtToken, err := rt.Generate(&authReq)
	assert.NoError(t, err)

	claims := jwt.MapClaims{}
	testToken, errToken := jwt.ParseWithClaims(jwtToken.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return rt.publicKey.CryptoPublicKey(), nil
	})
	assert.NoError(t, errToken)
	assert.True(t, testToken.Valid)
	assert.Equal(t, rt.tokenIssuer, claims["iss"])
	assert.Equal(t, authReq.Account, claims["sub"])
	assert.Equal(t, authReq.Service, claims["aud"])

}

func TestRegistryToken_CreateCerts(t *testing.T) {

	tmpDir, errDir := os.MkdirTemp("", "test_cert")
	require.NoError(t, errDir)

	tmpDir = filepath.ToSlash(tmpDir)

	rt := registryToken{}

	rt.serviceHost = "localhost"
	rt.serviceIP = "127.0.0.2"

	rt.RootPath = tmpDir
	rt.KeyPath = tmpDir + privateKeyName
	rt.PublicKeyPath = tmpDir + publicKeyName
	rt.CARootPath = tmpDir + CAName

	err := rt.createCerts()
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}()

	_, err = libtrust.LoadKeyFile(tmpDir + privateKeyName)
	assert.NoError(t, err)

	_, err = libtrust.LoadPublicKeyFile(tmpDir + publicKeyName)
	assert.NoError(t, err)

	_, err = libtrust.LoadCertificateBundle(tmpDir + CAName)
	assert.NoError(t, err)

	// test with error when certs exist
	err = rt.createCerts()
	assert.Error(t, err)
	assert.NoError(t, os.Remove(tmpDir+privateKeyName))

	err = rt.createCerts()
	assert.Error(t, err)
	assert.NoError(t, os.Remove(tmpDir+publicKeyName))

	err = rt.createCerts()
	assert.Error(t, err)
	assert.NoError(t, os.Remove(tmpDir+CAName))

	// test  with error when path error
	rt.Certs.KeyPath = "/"
	err = rt.createCerts()
	assert.Error(t, err)

	rt.Certs.PublicKeyPath = "/"
	err = rt.createCerts()
	assert.Error(t, err)

	rt.Certs.KeyPath = privateKeyName
	rt.Certs.PublicKeyPath = publicKeyName

	rt.Certs.CARootPath = "/"
	err = rt.createCerts()
	assert.Error(t, err)

}
