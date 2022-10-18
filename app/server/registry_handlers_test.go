package server

import (
	"bytes"
	"context"
	"fmt"
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

	// test without credentials
	request(t, "GET", "/api/v1/registry/auth", testRegistryHandlers.tokenAuth, nil, http.StatusUnauthorized)

	tests := []struct {
		login, password string
		query           string
		expectedStatus  int
	}{
		{
			// test with empty password
			login:          "test",
			password:       "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			// test with unknown user
			login:          "no_foo",
			password:       "foo_password",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			// test with bad user password
			login:          "foo",
			password:       "wrong_password",
			expectedStatus: http.StatusForbidden,
		},
		{
			// test with disabled user
			login:          "foo",
			password:       "password",
			expectedStatus: http.StatusForbidden,
		},
		{
			// test with no content
			login:          "bar",
			password:       "bar_password",
			expectedStatus: http.StatusNoContent,
		},
		{
			// test with login params
			login:          "bar",
			password:       "bar_password",
			query:          "?account=bar&client_id=docker&offline_token=true&service=container_registry",
			expectedStatus: http.StatusOK,
		},
		{
			// test with login params with wrong account name
			login:          "bar",
			password:       "bar_password",
			query:          "?account=test&client_id=docker&offline_token=true&service=container_registry",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			// test with resource fetch params
			login:          "bar",
			password:       "bar_password",
			query:          "?account=bar&scope=repository:test_resource_2:pull&service=container_registry",
			expectedStatus: http.StatusOK,
		},
		{
			// test with resource fetch params for user role
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_3:pull&service=container_registry",
			expectedStatus: http.StatusOK,
		},
		{
			// test with resource fetch params for user role with restricted scope
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_3:push&service=container_registry",
			expectedStatus: http.StatusForbidden,
		},
		{
			// test with resource fetch params for user role with bad scope
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository&service=container_registry",
			expectedStatus: http.StatusBadRequest,
		},
		{
			// test for token error with service name is unknown
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_3:pull&service=unknown_registry",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			// test with expire param
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_3:pull&service=container_registry&expire=3600",
			expectedStatus: http.StatusOK,
		},
		{
			// test with invalid expire param
			login:          "baz",
			password:       "baz_password",
			query:          "?account=baz&scope=repository:test_resource_3:pull&service=container_registry&expire=NaN",
			expectedStatus: http.StatusBadRequest,
		},
	}

	ctx := context.Background()
	for _, entry := range tests {
		t.Logf("test for entry: %v", entry)
		requestWithCredentials(t, ctx, entry.login, entry.password, "GET", fmt.Sprintf("/api/v1/registry/auth%s", entry.query), testRegistryHandlers.tokenAuth, nil, entry.expectedStatus)

	}
}

func TestRegistryHandlers_health(t *testing.T) {
	testRegistryHandlers := registryHandlers{}
	testRegistryHandlers.l = log.Default()

	testRegistryHandlers.registryService = prepareRegistryMock(t)
	testRegistryHandlers.dataStore = prepareAccessStoreMock(t)
	filledTestEntries(t, &testRegistryHandlers)

	ctx := context.Background()
	requestWithCredentials(t, ctx, "bar", "bar_password", "GET", "/api/v1/registry/health", testRegistryHandlers.health, nil, http.StatusOK)

	// test with error
	ctx = context.WithValue(ctx, ctxKey, false)
	requestWithCredentials(t, ctx, "bar", "bar_password", "GET", "/api/v1/registry/health", testRegistryHandlers.health, nil, http.StatusInternalServerError)
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
			"url": "https://example.com/v2/library/test/manifests/sha256:c3b3692957d439ac1928219a83fac91e7bf96c153725526874673ae1f2023f8d5"
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
		requestWithCredentials(t, ctx, "bar", "bar_password", "GET", "/api/v1/registry/events", testRegistryHandlers.events, test.body, test.expected)
	}
}

func TestRegistryHandlers_syncRepositories(t *testing.T) {
	testRegistryHandlers := registryHandlers{
		dataService: prepareDataServiceMock(),
	}
	testRegistryHandlers.l = log.Default()
	testRegistryHandlers.ctx = context.Background()

	requestWithCredentials(t, testRegistryHandlers.ctx, "bar", "bar_password", "GET", "/api/v1/registry/events", testRegistryHandlers.syncRepositories, nil, http.StatusOK)

	ctx := context.WithValue(testRegistryHandlers.ctx, ctxKey, errors.New("repo sync error"))
	testRegistryHandlers.ctx = ctx
	requestWithCredentials(t, ctx, "bar", "bar_password", "GET", "/api/v1/registry/events", testRegistryHandlers.syncRepositories, nil, http.StatusInternalServerError)
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

	return &registryInterfaceMock{
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
		ApiVersionCheckFunc: func(ctx context.Context) error {

			switch val := ctx.Value(ctxKey).(type) {
			case bool:
				if !val {
					return errors.New("failed to make health request")
				}
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

			switch val := id.(type) {
			case string:
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
			if len(filter.IDs) > 0 {
				for _, id := range filter.IDs {
					if val, ok := testAccesses[id]; ok {
						if val.Type == filter.Filters["resource_type"].(string) && val.ResourceName == filter.Filters["resource_name"].(string) && val.Action == filter.Filters["action"].([]string)[0] {
							testListResponse.Total++
							testListResponse.Data = append(testListResponse.Data, val)
						}
					}
				}
				if testListResponse.Total == 0 {
					return testListResponse, errors.New("access records not found")
				}
			}
			return testListResponse, nil
		},

		CreateRepositoryFunc: func(ctx context.Context, entry *store.RegistryEntry) error {
			return nil
		},

		FindRepositoriesFunc: func(ctx context.Context, filter engine.QueryFilter) (result engine.ListResponse, err error) {
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
func requestWithCredentials(t *testing.T, ctx context.Context, login, password string, method, url string, handler http.HandlerFunc, body []byte, expectedStatusCode int) *httptest.ResponseRecorder {

	req, errReq := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	require.NoError(t, errReq)

	req.SetBasicAuth(login, password)

	require.NoError(t, errReq)
	testWriter := httptest.NewRecorder()
	h := handler
	h.ServeHTTP(testWriter, req)
	assert.Equal(t, expectedStatusCode, testWriter.Code)
	return testWriter
}
