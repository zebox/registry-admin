package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSL_Redirect(t *testing.T) {
	s := Server{
		SSLConfig: SSLConfig{RedirHTTPPort: 8843},
	}

	ts := httptest.NewServer(s.httpToHTTPSRouter())
	defer ts.Close()

	client := http.Client{
		// prevent http redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},

		// allow self-signed certificate
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint
		},
	}

	// check http to https redirect response
	resp, err := client.Get(strings.Replace(ts.URL, "127.0.0.1", "localhost", 1) + "/blah?param=1")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, resp.Body.Close())
	}()
	assert.Equal(t, 307, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("https://localhost:%d/blah?param=1", s.SSLConfig.Port), resp.Header.Get("Location"))
}

func TestSSL_ACME_HTTPChallengeRouter(t *testing.T) {
	s := Server{
		SSLConfig: SSLConfig{
			Port:         chooseRandomUnusedPort(),
			ACMELocation: "acme",
			FQDNs:        []string{"example.com", "localhost"},
		},
	}

	m := s.makeAutocertManager()
	defer func() {
		assert.NoError(t, os.RemoveAll(s.SSLConfig.ACMELocation))
	}()

	ts := httptest.NewServer(s.httpChallengeRouter(m))
	defer ts.Close()

	client := http.Client{
		// prevent http redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	lh := strings.Replace(ts.URL, "127.0.0.1", "localhost", 1)
	// check http to https redirect response
	resp, err := client.Get(lh + "/blah?param=1")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, resp.Body.Close())
	}()

	assert.Equal(t, 307, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("https://localhost:%d/blah?param=1", s.SSLConfig.Port), resp.Header.Get("Location"))

	// check acme http challenge
	req, err := http.NewRequest("GET", lh+"/.well-known/acme-challenge/token123", http.NoBody)
	require.NoError(t, err)
	req.Host = "localhost" // for passing hostPolicy check
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, resp.Body.Close())
	}()
	assert.Equal(t, 404, resp.StatusCode)

	err = m.Cache.Put(context.Background(), "token123+http-01", []byte("token"))
	assert.NoError(t, err)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, resp.Body.Close())
	}()
	assert.Equal(t, 200, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "token", string(body))
}
