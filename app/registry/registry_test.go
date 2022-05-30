package registry

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
