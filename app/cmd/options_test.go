package cmd

import (
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testJSONConfig = `
{
  "listen": "127.0.0.1",
  "port": 8088,
  "host_name": "localhost",
  "auth": {
    "token_secret": "super-secret-test-token",
    "issuer_name": "test_issuer",
    "jwt_ttl": "10s",
    "cookie_ttl": "11"
  },
  "logger": {
    "stdout": true,
    "enabled": true,
    "file_name": "test_logger.log",
    "max_size": "100M",
    "max_backups": 2
  },
  "ssl": {
    "type": "none",
    "cert": "./cert/certificate.pem",
    "key": "./cert/privkey.pem",
    "acme_location":"./test_acme",
    "acme_email": "email@test.org",
    "acme_fqdns": ["test.org","demo.test.org"],
    "redir_http_port": 8443
  },
  "debug": true,
  "store": {
    "type": "embed",
    "embed": {
       "path": "./test.db"
    }
  }
}
`

var testYmlConfig = `
listen: 127.0.0.1
port: 8080

auth:
  token_secret: super-secret-test-token

store:
  embed:
    path: ./test.db
`

func Test_parseArgs(t *testing.T) {

	assert.NoError(t, os.Setenv("RA_LISTEN", "127.0.0.9"))
	assert.NoError(t, os.Setenv("RA_PORT", "9999"))

	// test for auth args
	assert.NoError(t, os.Setenv("RA_AUTH_TOKEN_SECRET", "test-super-token-secret"))
	assert.NoError(t, os.Setenv("RA_AUTH_HOST_NAME", "hostname.test"))
	assert.NoError(t, os.Setenv("RA_AUTH_ISSUER_NAME", "test-issuer"))
	assert.NoError(t, os.Setenv("RA_AUTH_JWT_TTL", "20s"))
	assert.NoError(t, os.Setenv("RA_AUTH_COOKIE_TTL", "30d"))

	// test for logger args
	assert.NoError(t, os.Setenv("RA_LOGGER_STDOUT", "true"))
	assert.NoError(t, os.Setenv("RA_LOGGER_ENABLED", "true"))
	assert.NoError(t, os.Setenv("RA_LOGGER_FILE", "./test_logger.log"))
	assert.NoError(t, os.Setenv("RA_LOGGER_SIZE", "999M"))
	assert.NoError(t, os.Setenv("RA_LOGGER_BACKUPS", "99"))

	// test for ssl args
	assert.NoError(t, os.Setenv("RA_SSL_TYPE", "none"))
	assert.NoError(t, os.Setenv("RA_SSL_CERT", "test.crt"))
	assert.NoError(t, os.Setenv("RA_SSL_KEY", "test.key"))
	assert.NoError(t, os.Setenv("RA_SSL_ACME_LOCATION", "./cert/path"))
	assert.NoError(t, os.Setenv("RA_SSL_ACME_EMAIL", "test@email.local"))
	assert.NoError(t, os.Setenv("RA_SSL_ACME_HTTP_PORT", "8080"))
	assert.NoError(t, os.Setenv("RA_SSL_PORT", "8433"))
	assert.NoError(t, os.Setenv("RA_SSL_ACME_FQDN", "test.domain.local"))

	// args for REGISTRY
	assert.NoError(t, os.Setenv("RA_REGISTRY_HOST", "test.registry-host.local"))
	assert.NoError(t, os.Setenv("RA_REGISTRY_PORT", "5000"))
	assert.NoError(t, os.Setenv("RA_REGISTRY_AUTH_TYPE", "basic"))

	// test for store args
	assert.NoError(t, os.Setenv("RA_STORE_DB_TYPE", "embed"))
	assert.NoError(t, os.Setenv("RA_STORE_EMBED_DB_PATH", "./db/data.db"))

	assert.NoError(t, os.Setenv("RA_DEBUG", "true"))

	var testMatcherOptions Options

	testMatcherOptions.Listen = "127.0.0.9"
	testMatcherOptions.Port = 9999
	testMatcherOptions.Auth.TokenSecret = "test-super-token-secret"
	testMatcherOptions.HostName = "localhost"
	testMatcherOptions.Auth.IssuerName = "test-issuer"
	testMatcherOptions.Auth.TokenDuration = "20s"
	testMatcherOptions.Auth.CookieDuration = "30d"

	testMatcherOptions.Logger.StdOut = true
	testMatcherOptions.Logger.Enabled = true
	testMatcherOptions.Logger.FileName = "./test_logger.log"
	testMatcherOptions.Logger.MaxSize = "999M"
	testMatcherOptions.Logger.MaxBackups = 99

	testMatcherOptions.SSL.Type = "none"
	testMatcherOptions.SSL.Cert = "test.crt"
	testMatcherOptions.SSL.Key = "test.key"
	testMatcherOptions.SSL.ACMELocation = "./cert/path"
	testMatcherOptions.SSL.ACMEEmail = "test@email.local"
	testMatcherOptions.SSL.RedirHTTPPort = 8080
	testMatcherOptions.SSL.Port = 8433
	testMatcherOptions.SSL.FQDNs = []string{"test.domain.local"}

	testMatcherOptions.Registry.Host = "test.registry-host.local"
	testMatcherOptions.Registry.Port = 5000
	testMatcherOptions.Registry.AuthType = "basic"

	testMatcherOptions.Store.Type = "embed"
	testMatcherOptions.Store.AdminPassword = "admin"
	testMatcherOptions.Store.Embed.Path = "./db/data.db"
	testMatcherOptions.Debug = true

	os.Args = []string{os.Args[0]} // clear Go test flags
	testOpts, errParse := ParseArgs()
	require.NoError(t, errParse)
	require.NotNil(t, testOpts)
	assert.Equal(t, &testMatcherOptions, testOpts)

	// test for random token generated for main auth token
	assert.NoError(t, os.Setenv("RA_AUTH_TOKEN_SECRET", ""))
	testOpts, errParse = ParseArgs()
	require.NoError(t, errParse)
	require.NotNil(t, testOpts)
	assert.NotEmpty(t, testOpts.Auth.TokenSecret)
}

func TestJsonConfigParser_ReadConfigFromFile(t *testing.T) {
	// create config test file
	f, errParse := os.CreateTemp(os.TempDir(), "test_config.json")
	require.NoError(t, errParse)

	defer func(path string) {
		assert.NoError(t, f.Close())
		errUnlink := syscall.Unlink(path)
		assert.NoError(t, errUnlink)
	}(f.Name())

	errParse = os.WriteFile(f.Name(), []byte(testJSONConfig), 0o0444)
	require.NoError(t, errParse)

	var (
		jcp         jsonConfigParser
		testOptions Options
	)

	errParse = jcp.ReadConfigFromFile(f.Name(), &testOptions)
	assert.NoError(t, errParse)
	assert.Equal(t, testOptions.Auth.TokenSecret, "super-secret-test-token")

	// test with fake file
	errParse = jcp.ReadConfigFromFile("unknown.file", &testOptions)
	assert.Error(t, errParse)
}

func TestYamlConfigParser_ReadConfigFromFile(t *testing.T) {
	// create config test file
	f, errParse := os.CreateTemp(os.TempDir(), "test_config.yml")
	require.NoError(t, errParse)

	defer func(path string) {
		assert.NoError(t, f.Close())
		errUnlink := syscall.Unlink(path)
		assert.NoError(t, errUnlink)
	}(f.Name())

	errParse = os.WriteFile(f.Name(), []byte(testYmlConfig), 0o0444)
	require.NoError(t, errParse)

	var (
		ycp         yamlConfigParser
		testOptions Options
	)

	errParse = ycp.ReadConfigFromFile(f.Name(), &testOptions)
	assert.NoError(t, errParse)
	assert.Equal(t, testOptions.Auth.TokenSecret, "super-secret-test-token")

	// test with fake file
	errParse = ycp.ReadConfigFromFile("unknown.file", &testOptions)
	assert.Error(t, errParse)
}
