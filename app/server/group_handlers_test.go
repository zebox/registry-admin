package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_groupCreateCtrl(t *testing.T) {
	testGroupHandlers := groupHandlers{}
	testGroupHandlers.dataStore = prepareGroupMock(t)

	group := store.Group{
		Name:        "test_group_1",
		Description: "test group_1 description",
	}

	{
		// testing for create a new group

		testResponse := responseMessage{}
		groupData, err := json.Marshal(group)
		require.NoError(t, err)

		req, errReq := http.NewRequest("POST", "/api/v1/groups", bytes.NewBuffer(groupData))
		require.NoError(t, errReq)
		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupCreateCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)

		err = json.NewDecoder(testWriter.Body).Decode(&testResponse)
		require.NoError(t, err)
		var groupAsMap map[string]interface{}
		err = json.Unmarshal(groupData, &groupAsMap)
		require.NoError(t, err)
		groupAsMap["id"] = float64(testResponse.ID)

		assert.Equal(t, testResponse.ID, int64(2))
		assert.Equal(t, testResponse.Data.(map[string]interface{}), groupAsMap)
	}
	{
		// testing for create a new group with error
		group.ID = 1
		groupData, err := json.Marshal(group)
		require.NoError(t, err)

		req, errReq := http.NewRequest("POST", "/api/v1/groups", bytes.NewBuffer(groupData))
		require.NoError(t, errReq)
		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupCreateCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

		// try request with body unmarshalling error
		req, errReq = http.NewRequest("POST", "/api/v1/groups", http.NoBody)
		require.NoError(t, errReq)
		testWriter = httptest.NewRecorder()
		handler = testGroupHandlers.groupCreateCtrl
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusInternalServerError, testWriter.Code)
	}
}

func Test_groupInfoCtrl(t *testing.T) {
	testGroupHandlers := groupHandlers{}
	testGroupHandlers.dataStore = prepareGroupMock(t)

	group := store.Group{
		Name:        "test_group_1",
		Description: "test group_1 description",
	}

	{
		// get existed group info by id
		groupData, err := json.Marshal(group)
		require.NoError(t, err)

		// create group first
		req, errReq := http.NewRequest("POST", "/api/v1/groups", bytes.NewBuffer(groupData))
		require.NoError(t, errReq)
		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupCreateCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)

		// try to receive existed group
		req, errReq = http.NewRequest("GET", "/api/v1/groups/2", http.NoBody)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "2")

		require.NoError(t, errReq)
		testWriter = httptest.NewRecorder()
		handler = testGroupHandlers.groupInfoCtrl
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)
	}
	{
		// try to get existed group info with error (group with id doesn't exist)
		req, errReq := http.NewRequest("GET", "/api/v1/groups/3", http.NoBody)
		require.NoError(t, errReq)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "3")

		require.NoError(t, errReq)
		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupInfoCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

		// try to get with wrong id type
		rctx = chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "abc")

		require.NoError(t, errReq)
		testWriter = httptest.NewRecorder()
		handler = testGroupHandlers.groupInfoCtrl
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusBadRequest, testWriter.Code)
	}
}

func Test_groupFindCtrl(t *testing.T) {
	testGroupHandlers := groupHandlers{}
	testGroupHandlers.dataStore = prepareGroupMock(t)

	// create test groups
	groups := []store.Group{
		{
			Name:        "test_group_1",
			Description: "test group_1 description",
		},
		{
			Name:        "test_group_2",
			Description: "test group_2 description",
		},
		{
			Name:        "test_group_3",
			Description: "test group_3 description",
		},
	}

	for _, g := range groups {

		groupData, err := json.Marshal(g)
		require.NoError(t, err)

		req, errReq := http.NewRequest("POST", "/api/v1/groups", bytes.NewBuffer(groupData))
		require.NoError(t, errReq)
		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupCreateCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)
	}

	{
		// fetch all existed group from store
		req, errReq := http.NewRequest("GET", `/api/v1/groups`, http.NoBody)
		require.NoError(t, errReq)

		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupFindCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)

		var testResponseList engine.ListResponse
		errBody := json.NewDecoder(testWriter.Body).Decode(&testResponseList)
		require.NoError(t, errBody)
		assert.Equal(t, int64(4), testResponseList.Total)
	}

	{
		// fetch all existed group from store
		req, errReq := http.NewRequest("GET", `/api/v1/groups?filter={"ids":[2,3]}`, http.NoBody)
		require.NoError(t, errReq)

		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupFindCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)

		var testResponseList engine.ListResponse
		errBody := json.NewDecoder(testWriter.Body).Decode(&testResponseList)
		require.NoError(t, errBody)
		assert.Equal(t, int64(2), testResponseList.Total)
	}

	{
		// fetch not existed group from store
		req, errReq := http.NewRequest("GET", `/api/v1/groups?filter={"ids":[88,99]}`, http.NoBody)
		require.NoError(t, errReq)

		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupFindCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusInternalServerError, testWriter.Code)
	}

	{
		// fetch group with filter error
		req, errReq := http.NewRequest("GET", `/api/v1/groups?filter={"ids":[ab,cd]}`, http.NoBody)
		require.NoError(t, errReq)

		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupFindCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusInternalServerError, testWriter.Code)
	}
}

func Test_groupUpdateCtrl(t *testing.T) {
	testGroupHandlers := groupHandlers{}
	testGroupHandlers.dataStore = prepareGroupMock(t)

	group := store.Group{
		Name:        "test_group_2",
		Description: "test group_2 description",
	}

	{
		// create group first
		groupData, err := json.Marshal(group)
		require.NoError(t, err)

		req, errReq := http.NewRequest("POST", "/api/v1/groups", bytes.NewBuffer(groupData))
		require.NoError(t, errReq)
		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupCreateCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)

	}
	{
		// try to update existed group
		group.Name = "updated_group_2"
		group.ID = 2

		groupData, err := json.Marshal(group)
		require.NoError(t, err)
		req, errReq := http.NewRequest("PUT", "/api/v1/groups/2", bytes.NewBuffer(groupData))
		require.NoError(t, errReq)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "2")

		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupUpdateCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)

		// checking group data for updated
		req, errReq = http.NewRequest("GET", "/api/v1/groups/2", http.NoBody)
		require.NoError(t, errReq)

		rctx = chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "2")

		require.NoError(t, errReq)
		testWriter = httptest.NewRecorder()
		handler = testGroupHandlers.groupInfoCtrl
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)

		var asMapGroupData map[string]interface{}
		err = json.Unmarshal(groupData, &asMapGroupData)
		require.NoError(t, err)

		var respMsg responseMessage
		errBodyRead := json.NewDecoder(testWriter.Body).Decode(&respMsg)
		require.NoError(t, errBodyRead)
		assert.Equal(t, asMapGroupData, respMsg.Data)
	}
	{
		// try to update with error
		group.ID = 1
		groupData, err := json.Marshal(group)
		require.NoError(t, err)

		// try with not existed group
		req, errReq := http.NewRequest("PUT", "/api/v1/groups/3", bytes.NewBuffer(groupData))
		require.NoError(t, errReq)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "3")

		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupUpdateCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

		// try with unparsed group id
		req, errReq = http.NewRequest("PUT", "/api/v1/groups/abc", bytes.NewBuffer(groupData))
		require.NoError(t, errReq)

		rctx = chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "abc")

		testWriter = httptest.NewRecorder()
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusBadRequest, testWriter.Code)

		// try with empty body
		req, errReq = http.NewRequest("PUT", "/api/v1/groups/2", http.NoBody)
		require.NoError(t, errReq)

		rctx = chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "2")

		testWriter = httptest.NewRecorder()
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

		// try with call Update func error
		req, errReq = http.NewRequest("PUT", "/api/v1/groups/1", bytes.NewBuffer(groupData))
		require.NoError(t, errReq)

		rctx = chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "1")

		testWriter = httptest.NewRecorder()
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusInternalServerError, testWriter.Code)
	}

}

func Test_groupDeleteCtrl(t *testing.T) {
	testGroupHandlers := groupHandlers{}
	testGroupHandlers.dataStore = prepareGroupMock(t)

	group := store.Group{
		Name:        "test_group_2",
		Description: "test group_2 description",
	}

	{
		// create group first
		groupData, err := json.Marshal(group)
		require.NoError(t, err)

		req, errReq := http.NewRequest("POST", "/api/v1/groups", bytes.NewBuffer(groupData))
		require.NoError(t, errReq)
		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupCreateCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)
	}
	{

		req, errReq := http.NewRequest("DELETE", "/api/v1/groups/2", http.NoBody)
		require.NoError(t, errReq)

		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "2")

		testWriter := httptest.NewRecorder()
		handler := http.HandlerFunc(testGroupHandlers.groupDeleteCtrl)
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusOK, testWriter.Code)

		// try to delete not existed group
		req, errReq = http.NewRequest("DELETE", "/api/v1/groups/2", http.NoBody)
		require.NoError(t, errReq)

		rctx = chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "2")

		testWriter = httptest.NewRecorder()
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusInternalServerError, testWriter.Code)

		// try to delete with wrong group id
		req, errReq = http.NewRequest("DELETE", "/api/v1/groups/abc", http.NoBody)
		require.NoError(t, errReq)

		rctx = chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rctx.URLParams.Add("id", "abc")

		testWriter = httptest.NewRecorder()
		handler.ServeHTTP(testWriter, req)
		assert.Equal(t, http.StatusBadRequest, testWriter.Code)
	}

}

func prepareGroupMock(t *testing.T) engine.Interface {
	testGroupStorage := make(map[int64]store.Group)

	var testGroupIndex int64
	testGroupIndex++ // increase for create default group item

	// suppose that default group create when database init at first time
	testGroupStorage[1] = store.Group{
		ID:          testGroupIndex,
		Name:        "testDefaultGroup",
		Description: "test default group description",
	}

	return &engine.InterfaceMock{
		CreateGroupFunc: func(ctx context.Context, group *store.Group) error {
			if group.ID != 0 {
				return fmt.Errorf("group with id [%d] not allowed for create", group.ID)
			}
			testGroupIndex++
			group.ID = testGroupIndex
			testGroupStorage[testGroupIndex] = *group

			return nil
		},

		GetGroupFunc: func(ctx context.Context, id int64) (group store.Group, err error) {
			if _, ok := testGroupStorage[id]; !ok {
				return group, fmt.Errorf("group with id [%d] not found", id)
			}
			return testGroupStorage[id], nil
		},

		FindGroupsFunc: func(ctx context.Context, filter engine.QueryFilter) (engine.ListResponse, error) {
			testListResponse := engine.ListResponse{}

			// fetch by ids
			if len(filter.IDs) > 0 {
				for _, id := range filter.IDs {
					if val, ok := testGroupStorage[id]; ok {
						testListResponse.Total++
						testListResponse.Data = append(testListResponse.Data, val)
					}
				}
				if testListResponse.Total == 0 {
					return testListResponse, errors.New("group records not found")
				}
				return testListResponse, nil
			}

			// fetch all records
			for _, v := range testGroupStorage {
				testListResponse.Total++
				testListResponse.Data = append(testListResponse.Data, v)
			}

			return testListResponse, nil
		},

		UpdateGroupFunc: func(ctx context.Context, group store.Group) error {
			if _, ok := testGroupStorage[group.ID]; !ok {
				return errors.Errorf("group with id [%d] not found", group.ID)
			}

			// for call error in test only. In real storage group can be to update
			if group.ID == 1 {
				return errors.New("default group can't be updated")
			}
			testGroupStorage[group.ID] = group
			return nil
		},

		DeleteGroupFunc: func(ctx context.Context, id int64) error {
			if _, ok := testGroupStorage[id]; !ok {
				return errors.Errorf("group with id [%d] not found", id)
			}
			delete(testGroupStorage, id)
			return nil
		},
	}
}
