package server

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func Test_userCreateCtrl(t *testing.T) {
	testUserHandlers := userHandlers{}
	testUserHandlers.dataStore = prepareUserMock(t)

	user := store.User{
		Login:    "test_login",
		Name:     "test_user",
		Password: "super-secret-password",
		Role:     "admin",
	}
	userData, err := json.Marshal(user)
	require.NoError(t, err)

	req, errReq := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userData))
	require.NoError(t, errReq)

	testWriter := httptest.NewRecorder()
	handler := http.HandlerFunc(testUserHandlers.userCreateCtrl)
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusOK, testWriter.Code)

	var testResponse responseMessage
	err = json.NewDecoder(testWriter.Body).Decode(&testResponse)
	require.NoError(t, err)
	user.ID = 1

	assert.Equal(t, testResponse.ID, int64(1))
	assert.Equal(t, testResponse.Data.(map[string]interface{})["login"], user.Login)

	// wrong user data
	user.Login = ""
	userData, err = json.Marshal(user)
	require.NoError(t, err)

	req, errReq = http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userData))
	require.NoError(t, errReq)

	testWriter = httptest.NewRecorder()
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

	user.ID = 1
	user.Login = "test_login"
	userData, err = json.Marshal(user)
	require.NoError(t, err)

	req, errReq = http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userData))
	require.NoError(t, errReq)

	testWriter = httptest.NewRecorder()
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

	badUserJson := `{"id":"0"}`
	req, errReq = http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer([]byte(badUserJson)))
	require.NoError(t, errReq)

	testWriter = httptest.NewRecorder()
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusInternalServerError, testWriter.Code)
}

func Test_userInfoCtrl(t *testing.T) {
	testUserHandlers := userHandlers{}
	testUserHandlers.dataStore = prepareUserMock(t)

	req, errReq := http.NewRequest("GET", "/api/v1/users/10001", http.NoBody)
	require.NoError(t, errReq)

	rctx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rctx.URLParams.Add("id", "10001")

	testWriter := httptest.NewRecorder()
	handler := http.HandlerFunc(testUserHandlers.userInfoCtrl)
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusOK, testWriter.Code)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rctx.URLParams.Add("id", "test_login")

	testWriter = httptest.NewRecorder()
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusBadRequest, testWriter.Code)

	// test with user not found or storage error
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rctx.URLParams.Add("id", "20001")

	testWriter = httptest.NewRecorder()
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

}

func Test_userUpdateCtrl(t *testing.T) {
	testUserHandlers := userHandlers{}
	testUserHandlers.dataStore = prepareUserMock(t)

	var user = store.User{
		ID:          10001,
		Login:       "test_login",
		Name:        "test_name2",
		Password:    "test_password2",
		Role:        "manager",
		Group:       888,
		Description: "user updated",
	}

	userData, err := json.Marshal(user)
	require.NoError(t, err)

	req, errReq := http.NewRequest("PUT", "/api/v1/users/10001", bytes.NewBuffer(userData))
	require.NoError(t, errReq)

	rctx := chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rctx.URLParams.Add("id", "10001")

	testWriter := httptest.NewRecorder()
	handler := http.HandlerFunc(testUserHandlers.userUpdateCtrl)
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusOK, testWriter.Code)

	{
		// test for update with error case
		user.ID = 20001
		userData, err := json.Marshal(user)
		require.NoError(t, err)

		req, errReq := http.NewRequest("PUT", "/api/v1/users/10002", bytes.NewBuffer(userData))
		require.NoError(t, errReq)
		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "10002")

		tWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testUserHandlers.userUpdateCtrl)
		handler.ServeHTTP(tWriter, req)
		assert.Equal(t, http.StatusInternalServerError, tWriter.Code)
	}

	{
		// test for json unmarshalling body error
		req, errReq := http.NewRequest("PUT", "/api/v1/users/10001", http.NoBody)
		require.NoError(t, errReq)
		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "10001")

		tWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testUserHandlers.userUpdateCtrl)
		handler.ServeHTTP(tWriter, req)
		assert.Equal(t, http.StatusInternalServerError, tWriter.Code)
	}

	req, errReq = http.NewRequest("GET", "/api/v1/users/10001", http.NoBody)
	require.NoError(t, errReq)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	testWriter = httptest.NewRecorder()
	handler = testUserHandlers.userInfoCtrl
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusOK, testWriter.Code)

	var response responseMessage
	err = json.NewDecoder(testWriter.Body).Decode(&response)
	require.NoError(t, err)
	require.NotNil(t, response.Data)

	var targetUser map[string]interface{}
	err = json.Unmarshal(userData, &targetUser)
	assert.NoError(t, err)
	targetUser["password"] = ""
	assert.Equal(t, response.Data, targetUser)

	// test with unknown id
	req, errReq = http.NewRequest("PUT", "/api/v1/users/-1", bytes.NewBuffer(userData))
	require.NoError(t, errReq)

	rctx = chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rctx.URLParams.Add("id", "-1")

	testWriter = httptest.NewRecorder()
	handler = testUserHandlers.userUpdateCtrl
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

	// test with unparsed id value
	rctx = chi.NewRouteContext()
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rctx.URLParams.Add("id", "not_parsed")

	testWriter = httptest.NewRecorder()
	handler = testUserHandlers.userUpdateCtrl
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusBadRequest, testWriter.Code)
}

func Test_userFindCtrl(t *testing.T) {
	testUserHandlers := userHandlers{}
	testUserHandlers.dataStore = prepareUserMock(t)

	req, errReq := http.NewRequest("GET", `/api/v1/users?filter={"ids":[10001,10002]}`, http.NoBody)
	require.NoError(t, errReq)

	testWriter := httptest.NewRecorder()
	handler := http.HandlerFunc(testUserHandlers.userFindCtrl)
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusOK, testWriter.Code)

	var result engine.ListResponse
	err := json.NewDecoder(testWriter.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)

	// test with error
	req, errReq = http.NewRequest("GET", `/api/v1/users?filter={"ids":["10001"]}`, http.NoBody)
	require.NoError(t, errReq)

	testWriter = httptest.NewRecorder()
	handler = testUserHandlers.userFindCtrl
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

	req, errReq = http.NewRequest("GET", `/api/v1/users?filter={"ids":[20001]}`, http.NoBody)
	require.NoError(t, errReq)

	testWriter = httptest.NewRecorder()
	handler = testUserHandlers.userFindCtrl
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusInternalServerError, testWriter.Code)
}

func Test_userDeleteCtr(t *testing.T) {
	testUserHandlers := userHandlers{}
	testUserHandlers.dataStore = prepareUserMock(t)

	req, errReq := http.NewRequest("GET", `/api/v1/users?filter={"ids":[10001]}`, http.NoBody)
	require.NoError(t, errReq)

	testWriter := httptest.NewRecorder()
	handler := http.HandlerFunc(testUserHandlers.userFindCtrl)
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusOK, testWriter.Code)

	{
		req, errReq = http.NewRequest("DELETE", `/api/v1/users/10001`, http.NoBody)
		require.NoError(t, errReq)

		// check item for exist
		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "10001")
		handler = testUserHandlers.userDeleteCtrl
		testWriter = httptest.NewRecorder()
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)

		// deleteDigest existed item
		req, errReq = http.NewRequest("DELETE", `/api/v1/users/wrong_id`, http.NoBody)
		require.NoError(t, errReq)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "wrong_id")
		testWriter = httptest.NewRecorder()
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusBadRequest, testWriter.Code)
	}

	{
		// try to delete user which already deleted item
		req, errReq = http.NewRequest("DELETE", `/api/v1/users/10001`, http.NoBody)
		require.NoError(t, errReq)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "20001")
		testWriter = httptest.NewRecorder()
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusInternalServerError, testWriter.Code)
	}

	req, errReq = http.NewRequest("GET", `/api/v1/users?filter={"ids":[10001]}`, http.NoBody)
	require.NoError(t, errReq)

	handler = testUserHandlers.userFindCtrl
	testWriter = httptest.NewRecorder()
	handler.ServeHTTP(testWriter, req)
	assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

}

func prepareUserMock(t *testing.T) *engine.InterfaceMock {
	var user = store.User{
		ID:          10001,
		Login:       "test_login",
		Name:        "test_name",
		Password:    "test_password",
		Role:        "user",
		Group:       90001,
		Description: "test user entry",
	}

	users := []store.User{
		{
			ID:          10001,
			Login:       "foo",
			Name:        "foo",
			Password:    "foo_password",
			Role:        "admin",
			Group:       1,
			Disabled:    true,
			Description: "foo_description",
		},
		{
			ID:          10002,
			Login:       "bar",
			Name:        "bar",
			Password:    "bar_password",
			Role:        "admin",
			Group:       1,
			Disabled:    false,
			Description: "bar_description",
		},
		{
			ID:          10003,
			Login:       "baz",
			Name:        "baz",
			Password:    "baz_password",
			Role:        "user",
			Group:       1,
			Disabled:    false,
			Description: "baz_description",
		},
		{
			ID:          10004,
			Login:       "qux",
			Name:        "qux",
			Password:    "qux_password",
			Role:        "manager",
			Group:       1,
			Disabled:    false,
			Description: "qux_description",
		},
	}

	return &engine.InterfaceMock{

		CreateUserFunc: func(ctx context.Context, user *store.User) error {
			if user.ID == 1 {
				return errors.New("user already exist with the same ID")
			}
			if !store.CheckRoleInList(user.Role) {
				return errors.New("unknown role")
			}
			user.ID = 1
			return nil
		},

		GetUserFunc: func(ctx context.Context, id interface{}) (store.User, error) {

			require.NoError(t, user.HashAndSalt())

			switch val := id.(type) {
			case string:

				if i, err := strconv.Atoi(val); err == nil && int64(i) == user.ID {
					return user, nil
				}
				if val == user.Login {
					return user, nil
				}
			case int, int64:
				if val == int64(10002) {
					newUser := user
					newUser.ID = 10002
					return newUser, nil
				}
				if val.(int64) == user.ID {
					return user, nil
				}
			default:
				return user, errors.New("unsupported val type")
			}
			return store.User{}, errors.New("user not found")
		},

		FindUsersFunc: func(ctx context.Context, filter engine.QueryFilter) (engine.ListResponse, error) {

			result := engine.ListResponse{}

			for _, user := range users {
				if filter.Filters != nil {
					if ids, ok := filter.Filters["ids"]; ok {

						for _, id := range ids.([]interface{}) {
							if v, ok := id.(float64); ok {
								if int64(v) == user.ID {
									result.Total += 1
									result.Data = append(result.Data, user)
								}
							}

						}
					}

				}

				if val, ok := filter.Filters["login"]; ok && val.(string) == user.Login {
					result.Total += 1
					result.Data = append(result.Data, user)
				}
			}

			if result.Total == 0 {
				return result, errors.New("users not found")
			}

			return result, nil
		},

		UpdateUserFunc: func(ctx context.Context, usr store.User) error {

			if usr.ID != 10001 {
				return errors.New("user not found")
			}
			user.Name = usr.Name

			if usr.Password != "" {
				user.Password = usr.Password
				assert.NoError(t, user.HashAndSalt())
			}

			user.Group = usr.Group
			user.Role = usr.Role
			user.Description = usr.Description

			return nil
		},

		DeleteUserFunc: func(ctx context.Context, id int64) error {
			removeItem := func(i int) []store.User {
				users[i] = users[len(users)-1]
				return users[:len(users)-1]
			}

			for i, user := range users {
				if user.ID == id {
					users = removeItem(i)
					return nil
				}
			}
			return errors.Errorf("user with id=%d not found", id)
		},

		DeleteAccessFunc: func(ctx context.Context, id int64) error {
			return nil
		},
	}
}
