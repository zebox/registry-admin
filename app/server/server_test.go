package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	log "github.com/go-pkgz/lgr"
	"github.com/zebox/gojwk/storage"
	"github.com/zebox/registry-admin/app/store/engine/embedded"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/token"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/gojwk"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
)

const (
	testPrivateKeyFileName = "test_private.key"
	testPublicKeyFileName  = "test_public.key"
)

type testCtxValue string

var testUserId = int64(rand.Uint32()) //nolint:gosec

func TestServer_RunNoneSSL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := Server{
		Listen:   "*",
		Port:     chooseRandomUnusedPort(),
		Hostname: "localhost",
		Authenticator: auth.NewService(auth.Opts{
			SecretReader: token.SecretFunc(func(aud string) (string, error) { return "secret", nil }),
			AvatarStore:  avatar.NewNoOp(),
		}),
		SSLConfig: SSLConfig{
			SSLMode: SSLNone,
		},
		Storage: prepareTestStorage(t),
	}

	go func() {
		assert.Equal(t, http.ErrServerClosed, srv.Run(ctx))
	}()

	go func() {
		<-ctx.Done()
		srv.Shutdown()
	}()

	waitForServerStart(srv.Port)

	client := http.Client{
		// prevent http redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	defer client.CloseIdleConnections()

	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/health", srv.Port))
	require.NoError(t, err)

	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "pong", string(body))

}

func TestServer_RunStaticSSL(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, dir, err := initTestKeys(ctx, t)

	require.NoError(t, err)
	_, err = os.Stat(dir + "/" + testPrivateKeyFileName)
	require.NoError(t, err)

	_, err = os.Stat(dir + "/" + testPublicKeyFileName)
	require.NoError(t, err)

	port := chooseRandomUnusedPort()

	srv := Server{
		Hostname: "localhost",
		Port:     port,
		Authenticator: auth.NewService(auth.Opts{
			AvatarStore: avatar.NewNoOp(),
		}),

		SSLConfig: SSLConfig{
			SSLMode:       SSLStatic,
			RedirHTTPPort: port,
			Port:          chooseRandomUnusedPort(),
			Key:           dir + "/" + testPrivateKeyFileName,
			Cert:          dir + "/CA_" + testPublicKeyFileName + ".crt",
		},
		Storage: prepareTestStorage(t),
	}

	go func() {
		assert.Equal(t, http.ErrServerClosed, srv.Run(ctx))
	}()

	waitForServerStart(srv.SSLConfig.Port)

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

	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/test?p=1", srv.Port))
	require.NoError(t, err)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("https://localhost:%d/test?p=1", srv.SSLConfig.Port), resp.Header.Get("Location"))

	resp, err = client.Get(fmt.Sprintf("https://localhost:%d/health", srv.SSLConfig.Port))
	require.NoError(t, err)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "pong", string(body))

	srv.Shutdown()
}

func TestServer__RunAutoSSL(t *testing.T) {

	port := chooseRandomUnusedPort()

	srv := Server{
		Hostname: "localhost",
		Port:     port,
		Authenticator: auth.NewService(auth.Opts{
			AvatarStore: avatar.NewNoOp(),
		}),

		SSLConfig: SSLConfig{
			SSLMode:       SSLAuto,
			RedirHTTPPort: port,
			Port:          chooseRandomUnusedPort(),
		},
		Storage: prepareTestStorage(t),
	}

	go func() {
		assert.Equal(t, http.ErrServerClosed, srv.Run(context.Background()))
	}()

	waitForServerStart(srv.SSLConfig.Port)

	client := http.Client{
		// prevent http redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	defer client.CloseIdleConnections()

	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/test?p=1", srv.Port))
	require.NoError(t, err)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("https://localhost:%d/test?p=1", srv.SSLConfig.Port), resp.Header.Get("Location"))

	srv.Shutdown()
}

func TestServer_ClaimUpdateFn(t *testing.T) {
	srv := Server{
		Storage: prepareTestStorage(t),
	}
	strUserId := strconv.FormatInt(testUserId, 10)
	var baseClaims token.Claims
	baseClaims.Id = strUserId
	baseClaims.User = &token.User{ID: strUserId, Name: "test_user"}

	extraClaims := srv.ClaimUpdateFn(baseClaims)
	assert.Equal(t, extraClaims.Id, strUserId)
	assert.Equal(t, extraClaims.User.Role, "user")

	baseClaims.User = nil // reset User claims
	baseClaims.Id = "-1"
	baseClaims.User = &token.User{ID: "-1", Name: "uknown"}

	extraClaims = srv.ClaimUpdateFn(baseClaims)
	assert.Equal(t, extraClaims.User.Role, "")

}

func TestServer_BasicAuthCheckerFn(t *testing.T) {

	srv := Server{
		Storage: prepareTestStorage(t),
	}

	{
		ok, claim, err := srv.BasicAuthCheckerFn("test_user", "test_password")

		assert.NoError(t, err)
		assert.True(t, ok)
		assert.Equal(t, claim.ID, strconv.FormatInt(testUserId, 10))
	}

	{
		ok, _, err := srv.BasicAuthCheckerFn("test_user", "fake_password")
		assert.Error(t, err)
		assert.False(t, ok)
	}

	{
		ok, _, err := srv.BasicAuthCheckerFn("fake_user", "fake_password")
		assert.Error(t, err)
		assert.False(t, ok)
	}

	{
		// test for disabled user
		var ctxValue testCtxValue = "user_disabled"
		srv.ctx = context.WithValue(context.Background(), ctxValue, true)
		ok, _, err := srv.BasicAuthCheckerFn("test_user", "test_password")
		assert.Error(t, err)
		assert.Equal(t, false, ok)
	}

}

func TestServer_Check(t *testing.T) {

	srv := Server{
		Storage: prepareTestStorage(t),
	}
	ok, err := srv.Check("test_user", "test_password")
	assert.NoError(t, err)
	assert.True(t, ok)

	ok, err = srv.Check("fake_user", "fake_password")
	assert.Error(t, err)
	assert.False(t, ok)
}
func TestServer_Validate(t *testing.T) {
	srv := Server{
		Storage: prepareTestStorage(t),
	}

	tc := token.Claims{User: nil}
	ok := srv.Validate("", tc)
	assert.False(t, ok)

	tokenUser := &token.User{
		ID:   "99999",
		Name: "test_user",
	}
	tc.User = tokenUser
	tc.User.SetBoolAttr("disabled", true)
	ok = srv.Validate("", tc)
	assert.False(t, ok)

	tc.User.SetBoolAttr("disabled", false)
	ok = srv.Validate("", tc)
	assert.True(t, ok)

}
func TestRest_Shutdown(t *testing.T) {
	srv := Server{
		Authenticator: &auth.Service{},
		Hostname:      "127.0.0.1",
		Port:          chooseRandomUnusedPort(),
		Storage:       prepareTestStorage(t),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// without waiting for channel close at the end goroutine will stay alive after test finish
	// which would create data race with next test
	go func() {
		<-ctx.Done()
		srv.Shutdown()
	}()

	st := time.Now()
	err := srv.Run(ctx)
	assert.Equal(t, err, http.ErrServerClosed)
	assert.True(t, time.Since(st).Seconds() < 1, "should take about 100ms")

}

func TestIntegrationUserOperationWithEmbeddedStore(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testStore := prepareTestDB(ctx, t)

	ts := prepareTestServer(ctx, t, testStore)
	waitForServerStart(ts.Port)

	{
		client := http.Client{
			// prevent http redirect
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		defer client.CloseIdleConnections()

		var cookies []*http.Cookie

		baseUrl := fmt.Sprintf("http://localhost:%d/api/v1", ts.Port)
		loginURL := fmt.Sprintf("http://localhost:%d/auth/test_local/login", ts.Port)

		{
			// try to make a request without auth
			resp, err := client.Get(baseUrl + "/users")
			require.NoError(t, err)
			defer func() { assert.NoError(t, resp.Body.Close()) }()
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		}
		loginFn := func(user, password string, statusCode int) error {
			// try to login request

			// credential for embedded store create when database file created
			// by default login: 'admin' password: admin
			resp, err := client.Get(loginURL + fmt.Sprintf("?user=%s&passwd=%s", user, password))
			if err != nil {
				return err
			}
			defer func() { assert.NoError(t, resp.Body.Close()) }()
			assert.Equal(t, statusCode, resp.StatusCode)
			cookies = resp.Cookies()
			return nil
		}

		assert.NoError(t, loginFn("admin", "admin", http.StatusOK))

		addCookiesFn := func(request *http.Request, cookies []*http.Cookie) {
			for _, c := range cookies {
				request.AddCookie(c)
			}
		}
		testUser := &store.User{
			Login:       "test_admin",
			Name:        "test_admin",
			Password:    "test_admin_password",
			Role:        "admin",
			Group:       1,
			Disabled:    false,
			Description: "test description for admin",
		}

		testUserData, e := json.Marshal(testUser)
		require.NoError(t, e)

		{
			// testing create user

			req, err := http.NewRequestWithContext(ctx, "POST", baseUrl+"/users", bytes.NewBuffer(testUserData))
			require.NoError(t, err)
			addCookiesFn(req, cookies)

			resp, errClient := client.Do(req)
			require.NoError(t, errClient)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// parsing a server response
			var respMsg responseMessage
			err = json.NewDecoder(resp.Body).Decode(&respMsg)
			require.NoError(t, err)
			defer func() { assert.NoError(t, resp.Body.Close()) }()

			createdUserData, errJson := json.Marshal(respMsg.Data)
			require.NoError(t, errJson)

			var createdUser store.User
			require.NoError(t, json.Unmarshal(createdUserData, &createdUser))
			testUser.ID = respMsg.ID
			pwdIsOk := store.ComparePassword(createdUser.Password, testUser.Password)
			assert.True(t, pwdIsOk)

			// reset password because hash string with salt may be not math
			testUser.Password = ""
			createdUser.Password = ""

			assert.Equal(t, *testUser, createdUser)

			// re-login under created user
			assert.NoError(t, loginFn("test_admin", "test_admin_password", http.StatusOK))
		}

		{
			// testing get user from store
			req, err := http.NewRequestWithContext(ctx, "GET", baseUrl+"/users/2", http.NoBody)
			require.NoError(t, err)
			addCookiesFn(req, cookies)

			resp, errClient := client.Do(req)
			require.NoError(t, errClient)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// parsing a server response
			var respMsg responseMessage
			err = json.NewDecoder(resp.Body).Decode(&respMsg)
			require.NoError(t, err)
			defer func() { assert.NoError(t, resp.Body.Close()) }()

			createdUserData, errJson := json.Marshal(respMsg.Data)
			require.NoError(t, errJson)

			var createdUser store.User
			require.NoError(t, json.Unmarshal(createdUserData, &createdUser))
			testUser.ID = respMsg.ID
			assert.Equal(t, *testUser, createdUser)
		}

		{
			// testing find users
			testUsers := []store.User{
				{
					Login:       "test_manager",
					Name:        "test_manager",
					Password:    "test_manager_password",
					Role:        "manager",
					Group:       1,
					Disabled:    false,
					Description: "test description for manager",
				},
				{
					Login:       "test_user",
					Name:        "test_user",
					Password:    "test_user_password",
					Role:        "user",
					Group:       1,
					Disabled:    false,
					Description: "test description for user",
				},
			}

			for _, u := range testUsers {
				uData, errJson := json.Marshal(u)
				require.NoError(t, errJson)
				req, err := http.NewRequestWithContext(ctx, "POST", baseUrl+"/users", bytes.NewBuffer(uData))
				require.NoError(t, err)
				addCookiesFn(req, cookies)

				resp, errClient := client.Do(req)
				require.NoError(t, errClient)
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}

			req, err := http.NewRequestWithContext(ctx, "GET", baseUrl+"/users", http.NoBody)
			require.NoError(t, err)
			addCookiesFn(req, cookies)

			resp, errClient := client.Do(req)
			require.NoError(t, errClient)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// parsing a server response
			var respMsg engine.ListResponse
			err = json.NewDecoder(resp.Body).Decode(&respMsg)
			require.NoError(t, err)
			defer func() { assert.NoError(t, resp.Body.Close()) }()
			assert.Equal(t, int64(4), respMsg.Total)
		}

		{
			// try to update user and disable user which issued token use
			testUser.Disabled = true
			testUser.Password = ""

			uData, errJson := json.Marshal(testUser)
			require.NoError(t, errJson)
			req, err := http.NewRequestWithContext(ctx, "PUT", baseUrl+"/users/2", bytes.NewBuffer(uData))
			require.NoError(t, err)
			addCookiesFn(req, cookies)

			resp, errClient := client.Do(req)
			require.NoError(t, errClient)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// try request after update when user is disabled
			req, err = http.NewRequestWithContext(ctx, "GET", baseUrl+"/users/2", http.NoBody)
			require.NoError(t, err)

			resp, errClient = client.Do(req)
			require.NoError(t, errClient)
			assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

			// try re-login with for disabled user
			assert.NoError(t, loginFn("test_admin", "test_admin_password", http.StatusInternalServerError))
		}
		{
			// try re-login with user role
			assert.NoError(t, loginFn("test_user", "test_user_password", http.StatusOK))

			req, err := http.NewRequestWithContext(ctx, "GET", baseUrl+"/users", http.NoBody)
			require.NoError(t, err)
			addCookiesFn(req, cookies)

			resp, errClient := client.Do(req)
			require.NoError(t, errClient)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		}
		{
			// try re-login with manager role
			assert.NoError(t, loginFn("test_manager", "test_manager_password", http.StatusOK))

			req, err := http.NewRequestWithContext(ctx, "GET", baseUrl+"/users", http.NoBody)
			require.NoError(t, err)
			addCookiesFn(req, cookies)

			resp, errClient := client.Do(req)
			require.NoError(t, errClient)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	}
	ts.Shutdown()
}

func waitForServerStart(port int) {
	// wait for up to 3 seconds for HTTPS server to start
	for i := 0; i < 300; i++ {
		time.Sleep(time.Millisecond * 10)
		conn, _ := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), time.Millisecond*10)
		if conn != nil {
			_ = conn.Close()
			break
		}
	}
}

// request is helper for testing handler request
func request(t *testing.T, method, url string, handler http.HandlerFunc, body []byte, expectedStatusCode int) *httptest.ResponseRecorder {

	req, errReq := http.NewRequest(method, url, bytes.NewBuffer(body))
	require.NoError(t, errReq)

	param := strings.Split(url, "/")
	if !strings.HasPrefix(url, "?") && len(param) > 4 {
		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", param[4])
	}

	require.NoError(t, errReq)
	testWriter := httptest.NewRecorder()
	h := http.HandlerFunc(handler)
	h.ServeHTTP(testWriter, req)
	assert.Equal(t, expectedStatusCode, testWriter.Code)
	return testWriter
}

func prepareTestStorage(t *testing.T) *engine.InterfaceMock {

	return &engine.InterfaceMock{
		GetUserFunc: func(ctx context.Context, id interface{}) (store.User, error) {

			testUser := store.User{
				ID:   testUserId,
				Name: "test_user",
				Role: "user",
			}

			if ctx != nil {
				var testCtxValue testCtxValue = "user_disabled"
				ctxValue := ctx.Value(testCtxValue)
				if ctxValue != nil && ctxValue.(bool) {
					testUser.Disabled = ctxValue.(bool)
				}
			}
			testUser.Password = "test_password"
			assert.NoError(t, testUser.HashAndSalt())

			switch val := id.(type) {
			case string:
				if i, err := strconv.Atoi(val); err == nil {
					if int64(i) != testUserId {
						return store.User{}, errors.New("unknown user id")
					}
					return testUser, nil
				}
				if strings.HasPrefix(val, "fake") || val != testUser.Name {
					return store.User{}, errors.New("failed to check user credentials")
				}
			case int, int64:
				iUserId := val.(int)
				if int64(iUserId) != testUserId {
					return store.User{}, errors.New("unknown user id")
				}
			}
			return testUser, nil
		},
		CloseFunc: func(ctx context.Context) error {
			return nil
		},
	}
}

func prepareTestServer(ctx context.Context, t *testing.T, storage engine.Interface) *Server {

	srv := Server{
		Listen:   "*",
		Port:     chooseRandomUnusedPort(),
		Hostname: "localhost",

		SSLConfig: SSLConfig{
			SSLMode: SSLNone,
		},
		Storage:   storage,
		AccessLog: nopWriteCloser{io.Discard},
		L:         log.Default(),
	}

	testAuthenticator := auth.NewService(auth.Opts{

		TokenDuration:  time.Minute,
		CookieDuration: time.Minute,
		Issuer:         "zebox tester",
		URL:            "http://localhost",
		SecureCookies:  true,
		DisableXSRF:    true,

		JWTQuery: "jwt",
		Logger:   log.Default(),

		SecretReader:     token.SecretFunc(func(aud string) (string, error) { return "test_secret", nil }),
		AvatarStore:      avatar.NewNoOp(),
		Validator:        &srv,
		ClaimsUpd:        token.ClaimsUpdFunc(srv.ClaimUpdateFn),
		BasicAuthChecker: srv.BasicAuthCheckerFn,
	})

	testAuthenticator.AddDirectProvider("test_local", &srv)
	srv.Authenticator = testAuthenticator

	go func() {
		assert.Equal(t, http.ErrServerClosed, srv.Run(ctx))
	}()

	return &srv
}

// prepareTestDB real embedded store for integration testing
func prepareTestDB(ctx context.Context, t *testing.T) *embedded.Embedded {
	testDBPath := os.TempDir() + "/test.db"

	_ = os.Remove(testDBPath)

	db := embedded.NewEmbedded(testDBPath)
	err := db.Connect(ctx)
	require.NoError(t, err)

	go func() {

		<-ctx.Done()
		err = db.Close(ctx)
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 50) // wait for close connection
		err = os.Remove(testDBPath)
		assert.NoError(t, err)

	}()

	return db
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

// initTestKeys will create self-signed test keys pair
func initTestKeys(ctx context.Context, t *testing.T) (keys *gojwk.Keys, dir string, err error) {

	dir, err = ioutil.TempDir(os.TempDir(), "tk")
	if err != nil {
		return nil, "", err
	}

	fileStore := storage.NewFileStorage(dir, testPrivateKeyFileName, testPublicKeyFileName)
	keys, _ = gojwk.NewKeys(gojwk.Storage(fileStore))

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{

			Organization:  []string{"OLYMP, INC."},
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

type nopWriteCloser struct{ io.Writer }

func (n nopWriteCloser) Close() error { return nil }
