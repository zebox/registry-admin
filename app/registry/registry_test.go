package registry

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net"
	"os"
	"testing"
)

func TestNewRegistry(t *testing.T) {

	tmpDir, errDir := os.MkdirTemp("", "test_cert")
	require.NoError(t, errDir)

	defer func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}()

	testSetting := Settings{
		AuthType: SelfToken,
		CertificatesPaths: Certs{
			RootPath:      tmpDir + "/" + certsDirName,
			KeyPath:       tmpDir + "/" + privateKeyName,
			PublicKeyPath: tmpDir + "/" + publicKeyName,
			CARootPath:    tmpDir + "/" + CAName,
		},
	}

	_, err := NewRegistry("test_login", "test_password", "test_secret", testSetting)
	require.NoError(t, err)

	// test with bad certs path
	testSetting.CertificatesPaths.KeyPath = "*"
	_, err = NewRegistry("test_login", "test_password", "test_secret", testSetting)
	require.Error(t, err)

	// test with empty secret
	_, err = NewRegistry("test_login", "test_password", "", testSetting)
	require.Error(t, err)

	// test with empty one of certs path fields
	testSetting.CertificatesPaths.PublicKeyPath = ""
	_, err = NewRegistry("test_login", "test_password", "test_secret", testSetting)
	require.Error(t, err)

	// test with empty last filed entry
	testSetting.CertificatesPaths = Certs{
		CARootPath: CAName,
	}
	_, err = NewRegistry("test_login", "test_password", "test_secret", testSetting)
	require.Error(t, err)

	// test with empty certs path
	testSetting.CertificatesPaths = Certs{}
	_, err = NewRegistry("test_login", "test_password", "test_secret", testSetting)
	require.NoError(t, err)

	// test with empty basic login
	testSetting.AuthType = Basic
	_, err = NewRegistry("", "test_password", "test_secret", testSetting)
	require.Error(t, err)

}

func TestRegistry_ApiCheck(t *testing.T) {
	testPort := chooseRandomUnusedPort()
	testRegistry := NewMockRegistry(t, "127.0.0.1", testPort, 0, 0)
	defer testRegistry.Close()

	r := Registry{settings: Settings{
		Host: "http://127.0.0.1",
		Port: testPort,
	}}

	apiError, err := r.ApiVersionCheck(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "", apiError.Message)
}

func TestRegistry_Catalog(t *testing.T) {
	testPort := chooseRandomUnusedPort()
	reposNumbers := 100
	tagsNumbers := 50
	testRegistry := NewMockRegistry(t, "127.0.0.1", testPort, reposNumbers, tagsNumbers)
	defer testRegistry.Close()

	r := Registry{settings: Settings{
		Host: "http://127.0.0.1",
		Port: testPort,
	}}

	repos, err := r.Catalog(context.Background(), "", "")
	assert.NoError(t, err)
	assert.Equal(t, reposNumbers, len(repos.List))

	// test with pagination
	repos, err = r.Catalog(context.Background(), "10", "")
	assert.NoError(t, err)
	assert.Equal(t, 10, len(repos.List))

}

func chooseRandomUnusedPort() (port int) {
	for i := 0; i < 10; i++ {
		port = 40000 + int(rand.Int31n(10000)) //nolint:gosec
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			_ = ln.Close()
			break
		}
	}
	return port
}

func parseUrlForNextLink(nextLink string) {

}
