package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/gojwk"
	"github.com/zebox/gojwk/storage"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"syscall"
	"testing"
	"time"
)

const (
	testPrivateKeyFileName = "test_private.key"
	testPublicKeyFileName  = "test_public.key"
)

func TestIntegrationMain(t *testing.T) {

	tmpHtpasswd, err := os.CreateTemp(os.TempDir(), "tmp")
	require.NoError(t, err)

	port := 40000 + int(rand.Int31n(10000)) //nolint:gosec // used in test only
	os.Args = []string{"test",
		"--listen=*", "--port=" + strconv.Itoa(port),
		"--auth.token-secret=super-secret", "--hostname=localhost", "--registry.host=http://test.registry-host.local",
		"--registry.htpasswd=" + tmpHtpasswd.Name(), "--registry.login=test_admin",
		"--debug", "--logger.enabled", "--logger.stdout", "--logger.file=" + os.TempDir() + "/registry-admin.log",
		"--ssl.type=none", "--store.type=embed", "--store.embed.path=" + os.TempDir() + "/test.db",
	}

	defer func() {
		t.Log("cleanup files")
		assert.NoError(t, os.Remove(os.TempDir()+"/registry-admin.log"))
		assert.NoError(t, os.Remove(os.TempDir()+"/test.db"))
	}()

	done := make(chan struct{})
	go func() {
		<-done
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.NoError(t, e)
	}()

	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
	}()

	// defer cleanup because require check below can fail
	defer func() {
		close(done)
		<-finished
	}()

	waitForHTTPServerStart(port)
	time.Sleep(time.Second)
	{
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/ping", port))
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "pong", string(body))
	}

	{
		// test for web static content from embed.FS
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d/empty.txt", port))
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, 200, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "empty web content", string(body))
	}
}

func TestMainWithSSLAndAuth(t *testing.T) {

	tmpHtpasswd, errTmp := ioutil.TempFile(os.TempDir(), "tmp")
	require.NoError(t, errTmp)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, dir, err := initTestCertKeys(ctx, t)

	require.NoError(t, err)
	keyFile := dir + "/" + testPrivateKeyFileName
	certFile := dir + "/CA_" + testPublicKeyFileName + ".crt"

	port := 40000 + int(rand.Int31n(10000))    //nolint:gosec // used in test only
	sslPort := 40000 + int(rand.Int31n(10000)) //nolint:gosec // used in test only
	os.Args = []string{"test",
		"--listen=*", "--port=" + strconv.Itoa(port),
		"--auth.token-secret=super-secret", "--hostname=localhost", "--registry.host=http://test.registry-host.local",
		"--debug", "--logger.enabled", "--logger.stdout", "--logger.file=" + os.TempDir() + "/registry-admin.log",
		"--registry.htpasswd=" + tmpHtpasswd.Name(), "--registry.login=test_admin",
		"--ssl.type=static", "--ssl.port=" + strconv.Itoa(sslPort), "--ssl.cert=" + certFile, "--ssl.key=" + keyFile, "--store.type=embed",
		"--store.embed.path=" + os.TempDir() + "/test.db",
	}

	defer func() {
		assert.NoError(t, os.Remove(os.TempDir()+"/registry-admin.log"))
		assert.NoError(t, os.Remove(os.TempDir()+"/test.db"))
	}()

	done := make(chan struct{})
	go func() {
		<-done
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.NoError(t, e)
	}()

	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
	}()

	// defer cleanup because require check below can fail
	defer func() {
		close(done)
		<-finished
	}()

	waitForHTTPServerStart(port)
	time.Sleep(time.Second)

	client := http.Client{
		// prevent http redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},

		// allow self-signed certificate
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}

	defer client.CloseIdleConnections()
	{
		resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port))
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	}
	{
		resp, err := client.Get(fmt.Sprintf("https://localhost:%d/ping", sslPort))
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, "pong", string(body))
	}

	{
		resp, err := client.Get(fmt.Sprintf("https://localhost:%d/auth/local/login?user=fakse&passwd=user", sslPort))
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		assert.NoError(t, err)
	}
	{
		resp, err := client.Get(fmt.Sprintf("https://localhost:%d/auth/local/login?user=admin&passwd=admin", sslPort))
		require.NoError(t, err)
		defer func() { assert.NoError(t, resp.Body.Close()) }()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NoError(t, err)
	}
}

/* func TestMain(m *testing.M) {
	// ignore is added only for GitHub Actions, can't reproduce locally
	goleak.VerifyTestMain(m)
} */

// initTestCertKeys will create self-signed test keys pair
func initTestCertKeys(ctx context.Context, t *testing.T) (keys *gojwk.Keys, dir string, err error) {

	dir, err = ioutil.TempDir(os.TempDir(), "tk")
	if err != nil {
		return nil, "", err
	}

	fileStore := storage.NewFileStorage(dir, testPrivateKeyFileName, testPublicKeyFileName)
	keys, _ = gojwk.NewKeys(gojwk.Storage(fileStore))

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{

			Organization:  []string{"TEST, INC."},
			Country:       []string{"SPC"},
			Province:      []string{""},
			Locality:      []string{"Mars"},
			StreetAddress: []string{"Mariner valley"},
			PostalCode:    []string{"000001"},
		},

		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, 1),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// add Subject Alternative Name for requested IP and Domain
	// it prevents untasted error with client request
	// https://oidref.com/2.5.29.17
	ca.IPAddresses = append(ca.IPAddresses, net.ParseIP("127.0.0.1"), net.ParseIP("::"))
	ca.DNSNames = append(ca.DNSNames, "localhost")

	// check keys for exist in the storage provider path
	if err = keys.Load(); err != nil {

		// if keys doesn't exist or load fail then create new
		if errGenerate := keys.Generate(); errGenerate != nil {
			return nil, "", errGenerate
		}

		// create CA certificate for created keys pair
		if errCreate := keys.CreateCAROOT(ca); errCreate != nil {
			return nil, "", errCreate
		}

		// if new keys pair created successfully save they to defined storage
		if errSave := keys.Save(); errSave != nil {
			return nil, "", errSave
		}

	}

	if errCreate := keys.CreateCAROOT(ca); errCreate != nil {
		return nil, "", errCreate
	}

	go func() {
		<-ctx.Done()
		assert.NoError(t, os.RemoveAll(dir))
	}()
	return keys, dir, nil
}

func waitForHTTPServerStart(port int) {
	// wait for up to 10 seconds for server to start before returning it
	client := http.Client{
		Timeout: time.Second,

		// allow self-signed certificate
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
	}
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 100)
		if resp, err := client.Get(fmt.Sprintf("http://localhost:%d/ping", port)); err == nil {
			_ = resp.Body.Close()
			return
		}
	}
}
