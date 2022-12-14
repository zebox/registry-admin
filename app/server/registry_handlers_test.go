package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-pkgz/auth/token"
	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"github.com/zebox/registry-admin/app/store/service"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type ctxKeyName string

var (
	ctxKey       ctxKeyName = "testCtxKey"
	testUsers               = make(map[int64]store.User)
	testAccesses            = make(map[int64]store.Access)
)

func TestRegistryHandlers_tokenAuth(t *testing.T) {
	testRegistryHandlers := registryHandlers{}
	testRegistryHandlers.l = log.Default()

	testRegistryHandlers.registryService = prepareRegistryMock(t)
	testRegistryHandlers.dataStore = prepareAccessStoreMock(t)

	filledTestEntries(t, &testRegistryHandlers)

	tests := []struct {
		name            string
		login, password string
		query           string
		expectedStatus  int
	}{
		{
			name:           "test with bad query params",
			login:          "",
			password:       "",
			query:          "?account;container_registry",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "test with access for valid user to public repositories (shared for all users)",
			login:          "",
			password:       "",
			query:          "?account=baz&scope=repository:test_resource_5:pull&service=container_registry",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "test with empty password without query params",
			login:          "test",
			password:       "",
			expectedStatus: http.StatusForbidden,
		},
		{

			name:           "test with empty password",
			login:          "test",
			password:       "",
			query:          "?account=baz&scope=repository:test_resource_5:pull&service=container_registry",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "test with unknown user",
			login:          "no_foo",
			password:       "foo_password",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "test with bad user password",
			login:          "foo",
			password:       "wrong_password",
			expectedStatus: http.StatusForbidden,
		},
		{

			name:           "test with disabled user",
			login:          "foo",
			password:       "password",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "test without params",
			login:          "bar",
			password:       "bar_password",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "test with login params",
			login:          "bar",
			password:       "bar_password",
			query:          "?account=bar&client_id=docker&offline_token=true&service=container_registry",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "test with login params with wrong account name",
			login:          "bar",
			password:       "bar_password",
			query:          "?account=test&client_id=docker&offline_token=true&service=container_registry",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "test with resource fetch params",
			login:          "bar",
			password:       "bar_password",
			query:          "?account=bar&scope=repository:test_resource_2:pull&service=container_registry",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "test with resource fetch params for user role",
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_3:pull&service=container_registry",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "test access for permit for only registered users",
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_6:pull&service=container_registry",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "test with resource fetch params for user role, but with denied scope",
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_3:push&service=container_registry",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "test with resource fetch params for user role with bad scope",
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository&service=container_registry",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "test for token error with service name is unknown",
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_3:pull&service=unknown_registry",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			// test with expire param
			name:           "test with expire param",
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_3:pull&service=container_registry&expire=3600",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "test with invalid expire param",
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_3:pull&service=container_registry&expire=NaN",
			expectedStatus: http.StatusBadRequest,
		},
	}

	ctx := context.Background()
	for _, entry := range tests {
		t.Logf("test entry: %v", entry.name)
		requestWithCredentials(ctx, t, entry.login, entry.password, "GET", fmt.Sprintf("/api/v1/registry/auth%s", entry.query), testRegistryHandlers.tokenAuth, nil, entry.expectedStatus)
	}
}

func TestRegistryHandlers_health(t *testing.T) {
	testRegistryHandlers := registryHandlers{}
	testRegistryHandlers.l = log.Default()

	testRegistryHandlers.registryService = prepareRegistryMock(t)
	testRegistryHandlers.dataStore = prepareAccessStoreMock(t)
	filledTestEntries(t, &testRegistryHandlers)

	ctx := context.Background()
	requestWithCredentials(ctx, t, "bar", "bar_password", "GET", "/api/v1/registry/health", testRegistryHandlers.health, nil, http.StatusOK)

	// test with error
	ctx = context.WithValue(ctx, ctxKey, false)
	requestWithCredentials(ctx, t, "bar", "bar_password", "GET", "/api/v1/registry/health", testRegistryHandlers.health, nil, http.StatusInternalServerError)
}

func TestRegistryHandlers_events(t *testing.T) {
	testEnvelope := `{
	"events": [
		{
		  "id": "asdf-asdf-asdf-asdf-0",
		  "timestamp": "2006-01-02T15:04:05Z",
		  "action": "pull",
		  "target": {
			"mediaType": "application/vnd.docker.distribution.manifest.v1+json",
			"length": 1,
			"digest": "sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			"repository": "library/test",
			"url": "https://example.com/v2/library/test/manifests/sha256:c3b3692957d439ac1928219a83fac91e7bf96c153725526874673ae1f2023f8d5",
			"references":[
			{
				"mediaType":"application/vnd.docker.container.image.v1+json",
				"digest":"sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf"
			}
			]
		  }
		}]
	}`

	testRegistryHandlers := registryHandlers{}
	testRegistryHandlers.l = log.Default()

	testRegistryHandlers.registryService = prepareRegistryMock(t)
	testRegistryHandlers.dataService = &service.DataService{Storage: prepareAccessStoreMock(t)}

	testsTable := []struct {
		name     string
		body     []byte
		expected int
	}{
		{
			name:     "working test",
			body:     []byte(testEnvelope),
			expected: http.StatusOK,
		},
		{
			name:     "with json unmarshal error",
			body:     nil,
			expected: http.StatusInternalServerError,
		},
		{
			name:     "with dataService error",
			body:     []byte(strings.Replace(testEnvelope, "pull", "unknown", 1)),
			expected: http.StatusInternalServerError,
		},
	}

	ctx := context.Background()
	for _, test := range testsTable {
		t.Log(test.name)
		requestWithCredentials(ctx, t, "bar", "bar_password", "GET", "/api/v1/registry/events", testRegistryHandlers.events, test.body, test.expected)
	}
}

func TestRegistryHandlers_syncRepositories(t *testing.T) {
	testRegistryHandlers := registryHandlers{
		dataService: prepareDataServiceMock(),
	}
	testRegistryHandlers.l = log.Default()
	testRegistryHandlers.ctx = context.Background()

	requestWithCredentials(testRegistryHandlers.ctx, t, "bar", "bar_password", "GET", "/api/v1/registry/events", testRegistryHandlers.syncRepositories, nil, http.StatusOK)

	ctx := context.WithValue(testRegistryHandlers.ctx, ctxKey, errors.New("repo sync error"))
	testRegistryHandlers.ctx = ctx
	requestWithCredentials(ctx, t, "bar", "bar_password", "GET", "/api/v1/registry/events", testRegistryHandlers.syncRepositories, nil, http.StatusInternalServerError)
}

func TestRegistryHandlers_imageConfig(t *testing.T) {
	testRegistryHandlers := registryHandlers{}
	testRegistryHandlers.l = log.Default()

	testRegistryHandlers.registryService = prepareRegistryMock(t)
	ctx := context.Background()
	testTable := []struct {
		name           string
		url            string
		ctx            context.Context
		expectedStatus int
	}{
		{
			name:           "request with empty name and digest params",
			url:            "/api/v1/registry/catalog/blobs",
			ctx:            ctx,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "successful request",
			url:            "/api/v1/registry/catalog/blobs?name=n_test&digest=d_test",
			ctx:            ctx,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "request with error response",
			url:            "/api/v1/registry/catalog/blobs?name=n_test&digest=d_test",
			ctx:            context.WithValue(ctx, ctxKey, true),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, test := range testTable {
		t.Log(test.name)
		requestWithCredentials(test.ctx, t, "bar", "bar_password", "GET", test.url, testRegistryHandlers.imageConfig, nil, test.expectedStatus)
	}
}

func TestRegistryHandlers_deleteDigest(t *testing.T) {
	testRegistryHandlers := registryHandlers{}
	testRegistryHandlers.l = log.Default()

	testRegistryHandlers.registryService = prepareRegistryMock(t)
	ctx := context.Background()
	testTable := []struct {
		name           string
		url            string
		ctx            context.Context
		expectedStatus int
	}{
		{
			name:           "request with empty name and digest params",
			url:            "/api/v1/registry/catalog/delete",
			ctx:            ctx,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "successful request",
			url:            "/api/v1/registry/catalog/delete?name=n_test&digest=d_test",
			ctx:            ctx,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "request with error response",
			url:            "/api/v1/registry/catalog/delete?name=n_test&digest=d_test",
			ctx:            context.WithValue(ctx, ctxKey, true),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, test := range testTable {
		t.Log(test.name)
		requestWithCredentials(test.ctx, t, "bar", "bar_password", "GET", test.url, testRegistryHandlers.deleteDigest, nil, test.expectedStatus)
	}
}

func TestRegistryHandlers_catalogList(t *testing.T) {

	testRegistryHandlers := registryHandlers{}
	testRegistryHandlers.l = log.Default()

	testRegistryHandlers.dataStore = prepareAccessStoreMock(t)
	filledTestEntries(t, &testRegistryHandlers)

	ctx := context.Background()
	testTable := []struct {
		name           string
		user           string
		url            string
		ctx            context.Context
		expectedStatus int
	}{
		{
			name:           "request with bad filter",
			user:           store.AdminRole,
			url:            `/api/v1/registry/catalog?&range=[0,A]`,
			ctx:            ctx,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "successful request with admin role",
			user:           store.AdminRole,
			url:            "/api/v1/registry/catalog",
			ctx:            ctx,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "successful request with user role",
			user:           store.UserRole,
			url:            "/api/v1/registry/catalog",
			ctx:            ctx,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "successful request with user role and filters",
			user:           store.UserRole,
			url:            "/api/v1/registry/catalog?filter={\"q\":\"test search\"}",
			ctx:            ctx,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "request with error response",
			user:           store.AdminRole,
			url:            "/api/v1/registry/catalog",
			ctx:            context.WithValue(ctx, ctxKey, true),
			expectedStatus: http.StatusInternalServerError,
		},
	}
	for _, test := range testTable {
		t.Log(test.name)
		requestWithCredentials(test.ctx, t, test.user, "bar_password", "GET", test.url, testRegistryHandlers.catalogList, nil, test.expectedStatus)
	}
}

func filledTestEntries(t *testing.T, testRegistryHandlers *registryHandlers) {
	var testUsersList = []store.User{
		{

			Login:       "foo",
			Name:        "foo",
			Password:    "foo_password",
			Role:        "admin",
			Group:       1,
			Disabled:    true,
			Description: "foo_description",
		},
		{

			Login:       "bar",
			Name:        "bar",
			Password:    "bar_password",
			Role:        "admin",
			Group:       1,
			Disabled:    false,
			Description: "bar_description",
		},
		{

			Login:       "baz",
			Name:        "baz",
			Password:    "baz_password",
			Role:        "user",
			Group:       2,
			Disabled:    false,
			Description: "baz_description",
		},
		{

			Login:       "qux",
			Name:        "qux",
			Password:    "qux_password",
			Role:        "manager",
			Group:       2,
			Disabled:    false,
			Description: "qux_description",
		},
	}

	var testAccessesList = []store.Access{
		{
			Name:         "test_access_1",
			Owner:        1,
			Type:         "repository",
			ResourceName: "test_resource_1",
			Action:       "pull:push",
		},
		{
			Name:         "test_access_2",
			Owner:        2,
			Type:         "repository",
			ResourceName: "test_resource_2",
			Action:       "push",
		},
		{
			Name:         "test_access_3",
			Owner:        3,
			Type:         "repository",
			ResourceName: "test_resource_3",
			Action:       "pull",
		},
		{
			Name:         "test_access_4",
			Owner:        4,
			Type:         "repository",
			ResourceName: "test_resource_4",
			Action:       "pull",
		},
		{
			Name:         "test_access_5",
			Owner:        engine.AnonymousUserID,
			Type:         "repository",
			ResourceName: "test_resource_5",
			Action:       "pull",
		},
		{
			Name:         "test_access_6",
			Owner:        engine.RegisteredUserID,
			Type:         "repository",
			ResourceName: "test_resource_6",
			Action:       "pull",
		},
	}

	ctx := context.Background()
	for _, user := range testUsersList {
		u := user
		err := testRegistryHandlers.dataStore.CreateUser(ctx, &u)
		assert.NoError(t, err)
		assert.NotEqual(t, int64(0), u.ID)

	}

	for _, access := range testAccessesList {
		a := access
		err := testRegistryHandlers.dataStore.CreateAccess(ctx, &a)
		assert.NoError(t, err)
		assert.NotEqual(t, int64(0), a.ID)

	}

}

func prepareRegistryMock(_ *testing.T) *registryInterfaceMock {
	if htpasswdMock == nil {
		htpasswdMock = make(map[string]struct{})
	}

	return &registryInterfaceMock{

		GetBlobFunc: func(ctx context.Context, name string, digest string) ([]byte, error) {
			if value := ctx.Value(ctxKey); value != nil {
				return nil, errors.New("failed to fetch image blob data")
			}
			return nil, nil
		},

		DeleteTagFunc: func(ctx context.Context, repoName string, digest string) error {
			if value := ctx.Value(ctxKey); value != nil {
				return errors.New("failed to delete image blob data")
			}
			return nil
		},
		LoginFunc: func(user store.User) (string, error) {
			if _, ok := testUsers[user.ID]; !ok {
				return "", errors.New("user not allowed here")
			}
			return "ok", nil
		},

		TokenFunc: func(authRequest registry.TokenRequest) (string, error) {
			for _, user := range testUsers {
				if user.Name == authRequest.Account && authRequest.Service != "unknown_registry" {
					return "ok", nil
				}
			}
			return "", errors.New("user not allowed here")
		},
		APIVersionCheckFunc: func(ctx context.Context) error {

			if val, ok := ctx.Value(ctxKey).(bool); ok {
				if !val {
					return errors.New("failed to make health request")
				}
			}

			return nil
		},

		UpdateHtpasswdFunc: func(usersFn registry.FetchUsers) error {
			if usersFn == nil {
				return errors.New("users list function should be defined")
			}

			users, err := usersFn.Users()
			if err != nil {
				return errors.New("users not found")
			}
			for _, u := range users {
				htpasswdMock[u.Login] = struct{}{}
			}
			return nil
		},
	}
}

func prepareAccessStoreMock(t *testing.T) *engine.InterfaceMock {

	var testUsersIndex, testAccessIndex int64

	return &engine.InterfaceMock{
		// ----------------------- USERS mock--------------------------------------
		CreateUserFunc: func(ctx context.Context, user *store.User) error {

			if _, ok := testUsers[user.ID]; ok {
				return errors.Errorf("user with id [%d] already exist", user.ID)
			}

			testUsersIndex++
			user.ID = testUsersIndex
			assert.NoError(t, user.HashAndSalt())
			testUsers[testUsersIndex] = *user
			return nil
		},

		GetUserFunc: func(ctx context.Context, id interface{}) (store.User, error) {

			if val, ok := id.(string); ok {
				for _, user := range testUsers {
					if val == user.Login {
						return user, nil
					}
				}
			}
			return store.User{}, fmt.Errorf("user with id [%d] not found", id)
		},

		// ----------------------- ACCESSES mock--------------------------------------

		CreateAccessFunc: func(ctx context.Context, access *store.Access) error {

			if _, ok := testAccesses[access.ID]; ok {
				return errors.Errorf("access with id [%d] already exist", access.ID)
			}

			testAccessIndex++
			access.ID = testAccessIndex
			testAccesses[testAccessIndex] = *access
			return nil
		},

		FindAccessesFunc: func(ctx context.Context, filter engine.QueryFilter) (engine.ListResponse, error) {
			testListResponse := engine.ListResponse{}

			// fetch by ids and filter values
			for _, val := range testAccesses {
				if val.Type == filter.Filters["resource_type"].(string) &&
					val.ResourceName == filter.Filters["resource_name"].(string) &&
					val.Action == filter.Filters["action"].([]string)[0] &&
					val.Owner == filter.Filters["owner_id"].(int64) {

					testListResponse.Total++
					testListResponse.Data = append(testListResponse.Data, val)
				}
			}
			if testListResponse.Total == 0 {
				return testListResponse, engine.ErrNotFound
			}
			return testListResponse, nil
		},

		CreateRepositoryFunc: func(ctx context.Context, entry *store.RegistryEntry) error {
			return nil
		},

		FindRepositoriesFunc: func(ctx context.Context, filter engine.QueryFilter) (result engine.ListResponse, err error) {
			if value := ctx.Value(ctxKey); value != nil {
				return result, errors.New("failed to get repository list")
			}

			if _, ok := filter.Filters[engine.RepositoriesByUserAccess]; ok {
				req, errReq := http.NewRequestWithContext(ctx, "GET", "https://test.local", http.NoBody)
				require.NoError(t, errReq)

				user, errUser := token.GetUserInfo(req)
				require.NoError(t, errUser)
				if user.Role != store.UserRole {
					return result, errors.New("owner id field should exist for 'user' role only")
				}
			}
			return result, err
		},
		UpdateRepositoryFunc: func(ctx context.Context, conditionClause map[string]interface{}, data map[string]interface{}) error {
			return nil
		},
	}
}

func prepareDataServiceMock() dataServiceInterface {
	return &dataServiceInterfaceMock{
		SyncExistedRepositoriesFunc: func(ctx context.Context) error {

			if value := ctx.Value(ctxKey); value != nil {
				return errors.New("sync repo error")
			}
			return nil
		},
	}
}

// requestWithCredentials is helper for testing handler request
//
//nolint:unparam // func used in tests only
func requestWithCredentials(ctx context.Context, t *testing.T, login, password, method, url string, handler http.HandlerFunc, body []byte, expectedStatusCode int) *httptest.ResponseRecorder {

	req, errReq := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	require.NoError(t, errReq)
	req.SetBasicAuth(login, password)

	// set user info to auth context
	req = token.SetUserInfo(req, token.User{
		Name: login,
		Role: login,
	})

	require.NoError(t, errReq)
	testWriter := httptest.NewRecorder()
	h := handler
	h.ServeHTTP(testWriter, req)
	assert.Equal(t, expectedStatusCode, testWriter.Code)
	return testWriter
}
