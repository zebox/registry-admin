package registry

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net"
	"net/url"
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
	var (
		total   int
		n, last string
	)
	n = "10"
	for total < reposNumbers {
		repos, err = r.Catalog(context.Background(), n, last)
		require.NoError(t, err)
		assert.Equal(t, 10, len(repos.List))
		total += len(repos.List)
		n, last, err = parseUrlForNextLink(repos.NextLink)
		require.NoError(t, err)
	}
	assert.Equal(t, reposNumbers, total)
}

func TestRegistry_ListingImageTags(t *testing.T) {
	testPort := chooseRandomUnusedPort()
	reposNumbers := 100
	tagsNumbers := 50
	testRegistry := NewMockRegistry(t, "127.0.0.1", testPort, reposNumbers, tagsNumbers)
	defer testRegistry.Close()

	r := Registry{settings: Settings{
		Host: "http://127.0.0.1",
		Port: testPort,
	}}

	for _, repoName := range testRegistry.repositories.List {
		tags, err := r.ListingImageTags(context.Background(), repoName, "", "")
		assert.NoError(t, err)
		assert.Equal(t, tagsNumbers, len(tags.Tags))
	}
	// test with pagination
	var (
		tags    ImageTags
		total   int
		n, last string
		err     error
	)
	n = "10"
	for _, repoName := range testRegistry.repositories.List {
		for total < reposNumbers*tagsNumbers {
			tags, err = r.ListingImageTags(context.Background(), repoName, n, last)
			require.NoError(t, err)
			assert.Equal(t, 10, len(tags.Tags))
			total += len(tags.Tags)
			n, last, err = parseUrlForNextLink(tags.NextLink)
			require.NoError(t, err)
		}
	}
	assert.Equal(t, reposNumbers*tagsNumbers, total)
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

func parseUrlForNextLink(nextLink string) (string, string, error) {
	result, err := url.ParseQuery(nextLink)
	if err != nil {
		return "", "", err
	}
	n := result.Get("n")
	last := result.Get("last")
	if n == "" && last == "" {
		return "", "", errors.New("page index is undefined in url params")
	}
	return n, last, nil
}
