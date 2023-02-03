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

	// test with defaults certs by auto create
	rt, err := NewRegistryToken()
	require.NoError(t, err)
	assert.Equal(t, int64(defaultTokenExpiration), rt.tokenExpiration)
	assert.Equal(t, defaultTokenIssuer, rt.tokenIssuer)
	assert.Equal(t, path+privateKeyName, rt.KeyPath)
	assert.Equal(t, path+publicKeyName, rt.PublicKeyPath)
	assert.Equal(t, path+caName, rt.CARootPath)

	// test with loading exist certs
	_, err = NewRegistryToken()
	require.NoError(t, err)

	// test with options
	tmpDir, errDir := os.MkdirTemp(os.TempDir(), "test_cert")
	require.NoError(t, errDir)
	rt, err = NewRegistryToken(

		TokenExpiration(10),
		TokenIssuer("127.0.0.2"),
		TokenLogger(log.Default()),
		CertsName(Certs{
			RootPath:      tmpDir,
			KeyPath:       tmpDir + "/test.key",
			PublicKeyPath: tmpDir + "/test.pub",
			CARootPath:    tmpDir + "/test_ca.crt",
			FQDNs:         []string{"domain.local.test"},
			IP:            "127.0.0.2",
		}),
	)

	require.NoError(t, err)
	assert.Equal(t, int64(10), rt.tokenExpiration)
	assert.Equal(t, "127.0.0.2", rt.tokenIssuer)
	assert.Equal(t, tmpDir+"/test.key", rt.KeyPath)
	assert.Equal(t, tmpDir+"/test.pub", rt.PublicKeyPath)
	assert.Equal(t, tmpDir+"/test_ca.crt", rt.CARootPath)
	assert.Equal(t, "127.0.0.2", rt.Certs.IP)
	assert.Contains(t, rt.Certs.FQDNs, "domain.local.test")

	// test with error
	_, err = NewRegistryToken(
		TokenExpiration(0))
	require.Error(t, err)

	// test with bad certs path
	badPath := "badPath"
	_, err = NewRegistryToken(
		CertsName(Certs{
			RootPath: badPath,
		}),
	)
	assert.Error(t, err)

	_, err = NewRegistryToken(
		CertsName(Certs{
			RootPath: tmpDir,
			KeyPath:  tmpDir + "/bad_key_path.key",
		}),
	)
	assert.Error(t, err)

	_, err = NewRegistryToken(
		CertsName(Certs{
			RootPath:      tmpDir,
			KeyPath:       tmpDir + "/test.key",
			PublicKeyPath: tmpDir + "/bad_public_path.pub",
		}),
	)
	assert.Error(t, err)

	_, err = NewRegistryToken(
		CertsName(Certs{
			RootPath:      tmpDir,
			KeyPath:       tmpDir + "/test.key",
			PublicKeyPath: tmpDir + "/test.pub",
			CARootPath:    tmpDir + "bac_ca_file_path.crt",
		}),
	)
	assert.Error(t, err)

	// with bad ip address
	_, err = NewRegistryToken(
		CertsName(Certs{
			RootPath:      tmpDir,
			KeyPath:       tmpDir + "/test.key",
			PublicKeyPath: tmpDir + "/test.pub",
			CARootPath:    tmpDir + "bac_ca_file_path.crt",
			IP:            "bad.ip.address",
		}),
	)
	assert.Error(t, err)
}

// Test for create registry token with a custom certificates, which created with an external tool
func TestToken_WithCustomTokenCerts(t *testing.T) {
	tmpDir, errDir := os.MkdirTemp(os.TempDir(), "test_cert")
	require.NoError(t, errDir)
	defer func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}()

	// for emit error when file loadings with empty content
	clearFileContentFn := func(path string) {
		f, errOpen := os.OpenFile(path, os.O_RDWR, 0o777)
		require.NoError(t, errOpen)
		require.NoError(t, f.Truncate(1))
		assert.NoError(t, f.Close())
	}

	testCerts := Certs{
		RootPath:      tmpDir,
		KeyPath:       tmpDir + "/test_private.key",
		PublicKeyPath: tmpDir + "/test_public.pub",
		CARootPath:    tmpDir + "/test_ca.crt",
		IP:            "127.0.0.1",
	}

	rt, err := NewRegistryToken(

		TokenExpiration(10),
		TokenIssuer("127.0.0.2"),
		TokenLogger(log.Default()),
		CertsName(testCerts),
	)

	require.NoError(t, err)

	err = rt.loadCerts()
	assert.NoError(t, err)

	// test for errors when certs files loads
	for _, cert := range []string{rt.Certs.CARootPath, rt.Certs.PublicKeyPath, rt.Certs.KeyPath} {
		clearFileContentFn(cert)
		assert.Error(t, rt.loadCerts())
	}

	_, err = NewRegistryToken(

		TokenExpiration(10),
		TokenIssuer("127.0.0.2"),
		TokenLogger(log.Default()),
		CertsName(testCerts),
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
		TokenIssuer("OLYMP TESTER"),
		TokenExpiration(3),
		CertsName(Certs{
			RootPath:      tmpDir,
			KeyPath:       tmpDir + "/test.key",
			PublicKeyPath: tmpDir + "/test.pub",
			CARootPath:    tmpDir + "/test_ca.crt",
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

	jwtToken, err := rt.generate(&authReq)
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

	rt := AccessToken{}

	rt.RootPath = tmpDir
	rt.KeyPath = tmpDir + privateKeyName
	rt.PublicKeyPath = tmpDir + publicKeyName
	rt.CARootPath = tmpDir + caName
	rt.Certs.FQDNs = []string{"localhost"}
	rt.Certs.IP = "127.0.0.2"

	err := rt.createCerts()
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}()

	_, err = libtrust.LoadKeyFile(tmpDir + privateKeyName)
	assert.NoError(t, err)

	_, err = libtrust.LoadPublicKeyFile(tmpDir + publicKeyName)
	assert.NoError(t, err)

	_, err = libtrust.LoadCertificateBundle(tmpDir + caName)
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
	assert.NoError(t, os.Remove(tmpDir+caName))

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
