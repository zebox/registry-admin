package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"syscall"
	"testing"
)

var testJsonConfig = `
{
  "listen": "127.0.0.1",
  "port": 8088,
  "authenticate": {
    "token_secret": "super-secret-test-token",
    "host_name": "localhost",
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

func TestParseArgs(t *testing.T) {
	{
		assert.NoError(t, os.Setenv("LISTEN", "127.0.0.9"))
		assert.NoError(t, os.Setenv("PORT", "9999"))

		// test for auth args
		assert.NoError(t, os.Setenv("AUTH_TOKEN_SECRET", "test-super-token-secret"))
		assert.NoError(t, os.Setenv("AUTH_HOST_NAME", "hostname.test"))
		assert.NoError(t, os.Setenv("AUTH_ISSUER_NAME", "test-issuer"))
		assert.NoError(t, os.Setenv("AUTH_JWT_TTL", "20s"))
		assert.NoError(t, os.Setenv("AUTH_COOKIE_TTL", "30d"))

		// test for logger args
		assert.NoError(t, os.Setenv("LOGGER_STDOUT", "true"))
		assert.NoError(t, os.Setenv("LOGGER_ENABLED", "true"))
		assert.NoError(t, os.Setenv("LOGGER_FILE", "./test_logger.log"))
		assert.NoError(t, os.Setenv("LOGGER_MAX_SIZE", "999M"))
		assert.NoError(t, os.Setenv("LOGGER_MAX_BACKUPS", "99"))

		// test for ssl args
		assert.NoError(t, os.Setenv("SSL_TYPE", "none"))
		assert.NoError(t, os.Setenv("SSL_CERT", "test.crt"))
		assert.NoError(t, os.Setenv("SSL_KEY", "test.key"))
		assert.NoError(t, os.Setenv("ACME_LOCATION", "./cert/path"))
		assert.NoError(t, os.Setenv("ACME_EMAIL", "test@email.local"))
		assert.NoError(t, os.Setenv("ACME_HTTP_PORT", "8080"))
		assert.NoError(t, os.Setenv("ACME_FQDN", "test.domain.local"))

		// test for store args
		assert.NoError(t, os.Setenv("DB_TYPE", "embed"))
		assert.NoError(t, os.Setenv("DB_PATH", "./db/path"))

		assert.NoError(t, os.Setenv("DEBUG", "true"))
	}
	var testMatcherOptions Options
	{
		testMatcherOptions.Listen = "127.0.0.9"
		testMatcherOptions.Port = 9999
		testMatcherOptions.Auth.TokenSecret = "test-super-token-secret"
		testMatcherOptions.HostName = "hostname.test"
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
		testMatcherOptions.SSL.FQDNs = []string{"test.domain.local"}

		testMatcherOptions.Store.Type = "embed"
		testMatcherOptions.Store.Embed.Path = "./db/path"
		testMatcherOptions.Debug = true
	}
}

func TestJsonConfigParser_ReadConfigFromFile(t *testing.T) {
	// create config test file
	f, err := ioutil.TempFile("", "test_config.json")
	require.NoError(t, err)

	defer func(path string) {
		assert.NoError(t, f.Close())
		errUnlink := syscall.Unlink(path)
		assert.NoError(t, errUnlink)
	}(f.Name())

	err = ioutil.WriteFile(f.Name(), []byte(testJsonConfig), 0644)
	require.NoError(t, err)

	var (
		jcp         jsonConfigParser
		testOptions Options
	)

	err = jcp.ReadConfigFromFile(f.Name(), &testOptions)
	assert.NoError(t, err)
	assert.Equal(t, testOptions.Auth.TokenSecret, "super-secret-test-token")

	// test wuth fake file
	err = jcp.ReadConfigFromFile("unknown.file", &testOptions)
	assert.Error(t, err)
}
