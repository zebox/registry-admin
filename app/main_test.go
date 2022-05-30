package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"github.com/shirou/gopsutil/process"
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

func Test_Main(t *testing.T) {

	port := 40000 + int(rand.Int31n(10000)) //nolint:gosec
	os.Args = []string{"test",
		"--listen=*", "--port=" + strconv.Itoa(port),
		"--auth.token-secret=super-secret", "--auth.hostname=localhost",
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
		t.Log("done processing")
		p, err := process.NewProcess(int32(syscall.Getpid()))
		require.NoError(t, err)

		childrenProccess, errChild := p.Children()
		require.NoError(t, errChild)
		t.Log("process killing")
		for _, v := range childrenProccess {
			errKill := v.Kill()
			t.Log(errKill)
			assert.NoError(t, errKill)

		}

		t.Log("process parent killing")
		errKillParent := p.Kill()
		assert.NoError(t, errKillParent) // Kill the parent process

		t.Log("processes killed")
	}()

	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
		t.Log("main finished")
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
}

func Test_MainWithSSLAndAuth(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, dir, err := initTestCertKeys(ctx, t)

	require.NoError(t, err)
	keyFile := dir + "/" + testPrivateKeyFileName
	certFile := dir + "/CA_" + testPublicKeyFileName + ".crt"

	port := 40000 + int(rand.Int31n(10000))    //nolint:gosec
	sslPort := 40000 + int(rand.Int31n(10000)) //nolint:gosec
	os.Args = []string{"test",
		"--listen=*", "--port=" + strconv.Itoa(port),
		"--auth.token-secret=super-secret", "--auth.hostname=localhost",
		"--debug", "--logger.enabled", "--logger.stdout", "--logger.file=" + os.TempDir() + "/registry-admin.log",
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
		p, err := process.NewProcess(int32(syscall.Getpid()))
		require.NoError(t, err)

		childrenProccess, errChild := p.Children()
		require.NoError(t, errChild)
		for _, v := range childrenProccess {
			assert.NoError(t, v.Kill())
		}
		assert.NoError(t, p.Kill()) // Kill the parent process

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

//initTestKeys will create self-signed test keys pair
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
	ca.IPAddresses = append(ca.IPAddresses, net.ParseIP("127.0.0.1"))
	ca.IPAddresses = append(ca.IPAddresses, net.ParseIP("::"))
	ca.DNSNames = append(ca.DNSNames, "localhost")

	// check keys for exist in the storage provider path
	if err = keys.Load(); err != nil {

		// if keys doesn't exist or load fail then create new
		if err = keys.Generate(); err != nil {
			return nil, "", err
		}

		// create CA certificate for created keys pair
		if err = keys.CreateCAROOT(ca); err != nil {
			return nil, "", err
		}

		// if new keys pair created successfully save they to defined storage
		if err = keys.Save(); err != nil {
			return nil, "", err
		}

	}

	if err = keys.CreateCAROOT(ca); err != nil {
		return nil, "", err
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
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
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
