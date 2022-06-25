package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	testUsers    = make(map[int64]store.User)
	testAccesses = make(map[int64]store.Access)
)

func Test_tokenAuth(t *testing.T) {
	testRegistryHandlers := registryHandlers{}
	testRegistryHandlers.l = log.Default()

	testRegistryHandlers.registryService = prepareRegistryMock()
	testRegistryHandlers.dataStore = prepareUserAccessStoreMock(t)

	filledTestEntries(t, &testRegistryHandlers)

	// test without credentials
	request(t, "GET", "/api/v1/registry/auth", testRegistryHandlers.tokenAuth, nil, http.StatusUnauthorized)

	// test with empty password
	requestWithCredentials(t, "test_login", "", "GET", "/api/v1/registry/auth", testRegistryHandlers.tokenAuth, nil, http.StatusUnauthorized)

	// test with unknown user
	requestWithCredentials(t, "no_foo", "foo_password", "GET", "/api/v1/registry/auth", testRegistryHandlers.tokenAuth, nil, http.StatusUnauthorized)

	// test with bad user password
	requestWithCredentials(t, "foo", "wrong_password", "GET", "/api/v1/registry/auth", testRegistryHandlers.tokenAuth, nil, http.StatusForbidden)

	// test with disabled user
	requestWithCredentials(t, "foo", "foo_password", "GET", "/api/v1/registry/auth", testRegistryHandlers.tokenAuth, nil, http.StatusForbidden)

	// test with no content
	requestWithCredentials(t, "bar", "bar_password", "GET", "/api/v1/registry/auth", testRegistryHandlers.tokenAuth, nil, http.StatusNoContent)

	// test with login params
	requestWithCredentials(t, "bar", "bar_password", "GET", "/api/v1/registry/auth?account=bar&client_id=docker&offline_token=true&service=container_registry", testRegistryHandlers.tokenAuth, nil, http.StatusOK)

	// test with login params with wrong account name
	requestWithCredentials(t, "bar", "bar_password", "GET", "/api/v1/registry/auth?account=test&client_id=docker&offline_token=true&service=container_registry", testRegistryHandlers.tokenAuth, nil, http.StatusInternalServerError)

	// test with resource fetch params
	requestWithCredentials(t, "bar", "bar_password", "GET", "/api/v1/registry/auth?account=bar&scope=repository:test_resource_2:pull&service=container_registry", testRegistryHandlers.tokenAuth, nil, http.StatusOK)

	// test with resource fetch params for user role
	requestWithCredentials(t, "baz", "baz_password", "GET", "/api/v1/registry/auth?account=baz&scope=repository:test_resource_3:pull&service=container_registry", testRegistryHandlers.tokenAuth, nil, http.StatusOK)

	// test with resource fetch params for user role with restricted scope
	requestWithCredentials(t, "baz", "baz_password", "GET", "/api/v1/registry/auth?account=baz&scope=repository:test_resource_3:push&service=container_registry", testRegistryHandlers.tokenAuth, nil, http.StatusForbidden)

	// test with resource fetch params for user role with bad scope
	requestWithCredentials(t, "baz", "baz_password", "GET", "/api/v1/registry/auth?account=baz&scope=repository&service=container_registry", testRegistryHandlers.tokenAuth, nil, http.StatusBadRequest)

	// test for token error
	requestWithCredentials(t, "baz", "baz_password", "GET", "/api/v1/registry/auth?account=baz&scope=repository:test_resource_3:pull&service=unknown_registry", testRegistryHandlers.tokenAuth, nil, http.StatusInternalServerError)

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

func prepareRegistryMock() *registryInterfaceMock {

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
	}
}

func prepareUserAccessStoreMock(t *testing.T) *engine.InterfaceMock {

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
	}
}

// requestWithCredentials is helper for testing handler request
func requestWithCredentials(t *testing.T, login, password string, method, url string, handler http.HandlerFunc, body []byte, expectedStatusCode int) *httptest.ResponseRecorder {

	req, errReq := http.NewRequest(method, url, bytes.NewBuffer(body))
	require.NoError(t, errReq)

	req.SetBasicAuth(login, password)

	param := strings.Split(url, "/")
	if !strings.HasPrefix(url, "?") && len(param) > 4 {
		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", param[4])
	}

	require.NoError(t, errReq)
	testWriter := httptest.NewRecorder()
	h := handler
	h.ServeHTTP(testWriter, req)
	assert.Equal(t, expectedStatusCode, testWriter.Code)
	return testWriter
}
