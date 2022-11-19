package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"net/http"
	"testing"
)

func Test_accessAddCtrl(t *testing.T) {
	testAccessHandlers := accessHandlers{}
	testAccessHandlers.dataStore = prepareAccessMock()

	access := store.Access{
		Name:         "test_access_1",
		Owner:        1,
		Type:         "registry",
		ResourceName: "test_resource",
		Action:       "pull:push",
	}

	{
		// testing for add a new access
		testResponse := responseMessage{}
		accessData, err := json.Marshal(access)
		require.NoError(t, err)

		testWriter := request(t, "POST", "/api/v1/access", testAccessHandlers.accessAddCtrl, accessData, http.StatusOK)

		err = json.NewDecoder(testWriter.Body).Decode(&testResponse)
		require.NoError(t, err)
		var accessAsMap map[string]interface{}
		err = json.Unmarshal(accessData, &accessAsMap)
		require.NoError(t, err)
		accessAsMap["id"] = float64(testResponse.ID)

		assert.Equal(t, testResponse.ID, int64(1))
		assert.Equal(t, testResponse.Data.(map[string]interface{}), accessAsMap)
	}
	{
		// testing for create new access with error
		access.ID = 1
		accessData, err := json.Marshal(access)
		require.NoError(t, err)

		request(t, "POST", "/api/v1/access", testAccessHandlers.accessAddCtrl, accessData, http.StatusInternalServerError)

		// try request with body unmarshalling error
		request(t, "POST", "/api/v1/access", testAccessHandlers.accessAddCtrl, nil, http.StatusInternalServerError)

	}
}

func Test_accessInfoCtrl(t *testing.T) {
	testAccessHandlers := accessHandlers{}
	testAccessHandlers.dataStore = prepareAccessMock()

	access := store.Access{
		Name:         "test_access_1",
		Owner:        1,
		Type:         "registry",
		ResourceName: "test_resource",
		Action:       "pull:push",
	}

	accessData, err := json.Marshal(access)
	require.NoError(t, err)

	// create access first
	request(t, "POST", "/api/v1/access", testAccessHandlers.accessAddCtrl, accessData, http.StatusOK)

	// try to receive existed access
	request(t, "GET", "/api/v1/access/1", testAccessHandlers.accessInfoCtrl, nil, http.StatusOK)

	// try to get existed access info with error (access with id doesn't exist)
	request(t, "GET", "/api/v1/access/2", testAccessHandlers.accessInfoCtrl, nil, http.StatusInternalServerError)

	// try to get with wrong id type
	request(t, "GET", "/api/v1/access/abc", testAccessHandlers.accessInfoCtrl, nil, http.StatusBadRequest)

}

func Test_accessFindCtrl(t *testing.T) {
	testAccessHandlers := accessHandlers{}
	testAccessHandlers.dataStore = prepareAccessMock()

	accesses := []store.Access{
		{
			Name:         "test_access_1",
			Owner:        1,
			Type:         "registry",
			ResourceName: "test_resource_1",
			Action:       "pull:push",
		},
		{
			Name:         "test_access_2",
			Owner:        1,
			Type:         "registry",
			ResourceName: "test_resource_2",
			Action:       "push",
		},
		{
			Name:         "test_access_3",
			Owner:        1,
			Type:         "*",
			ResourceName: "test_resource_3",
			Action:       "pull",
		},
	}
	for _, a := range accesses {
		accessData, err := json.Marshal(a)
		require.NoError(t, err)
		request(t, "POST", "/api/v1/access", testAccessHandlers.accessAddCtrl, accessData, http.StatusOK)
	}

	{
		// try to fetch all access
		testWriter := request(t, "GET", "/api/v1/access", testAccessHandlers.accessFindCtrl, nil, http.StatusOK)
		var responseList engine.ListResponse
		err := json.NewDecoder(testWriter.Body).Decode(&responseList)
		require.NoError(t, err)
		assert.Equal(t, int64(3), responseList.Total)
	}
	{
		testWriter := request(t, "GET", `/api/v1/access?filter={"ids":[2,3]}`, testAccessHandlers.accessFindCtrl, nil, http.StatusOK)
		var responseList engine.ListResponse
		err := json.NewDecoder(testWriter.Body).Decode(&responseList)
		require.NoError(t, err)
		assert.Equal(t, int64(2), responseList.Total)
	}

	request(t, "GET", `/api/v1/access?filter={"ids":[a,b]}`, testAccessHandlers.accessFindCtrl, nil, http.StatusInternalServerError)
	request(t, "GET", `/api/v1/access?filter={"ids":[4]}`, testAccessHandlers.accessFindCtrl, nil, http.StatusInternalServerError)
}

func Test_accessUpdateCtrl(t *testing.T) {
	testAccessHandlers := accessHandlers{}
	testAccessHandlers.dataStore = prepareAccessMock()

	access := store.Access{
		Name:         "test_access_1",
		Owner:        1,
		Type:         "registry",
		ResourceName: "test_resource",
		Action:       "pull:push",
	}

	// create access first
	accessData, err := json.Marshal(access)
	require.NoError(t, err)

	resp := request(t, "POST", "/api/v1/access", testAccessHandlers.accessAddCtrl, accessData, http.StatusOK)
	var responseMsg responseMessage
	err = json.NewDecoder(resp.Body).Decode(&responseMsg)
	require.NoError(t, err)

	access.ID = responseMsg.ID

	{ // try to update existed access
		access.Name = "updated_access_2"

		accessData, err = json.Marshal(access)
		require.NoError(t, err)
		request(t, "PUT", "/api/v1/access/1", testAccessHandlers.accessUpdateCtrl, accessData, http.StatusOK)

		// checking access data for updated
		testWriter := request(t, "GET", "/api/v1/access/1", testAccessHandlers.accessInfoCtrl, accessData, http.StatusOK)

		var asMapAccessData map[string]interface{}
		err = json.Unmarshal(accessData, &asMapAccessData)
		require.NoError(t, err)

		var respMsg responseMessage
		errBodyRead := json.NewDecoder(testWriter.Body).Decode(&respMsg)
		require.NoError(t, errBodyRead)
		assert.Equal(t, asMapAccessData, respMsg.Data)
	}

	{
		// try to update with error
		access.ID = -1
		accessData, err := json.Marshal(access)
		require.NoError(t, err)

		// try with not existed access
		request(t, "PUT", "/api/v1/access/99", testAccessHandlers.accessUpdateCtrl, accessData, http.StatusInternalServerError)

		// try with unparsed access id
		request(t, "PUT", "/api/v1/access/abc", testAccessHandlers.accessUpdateCtrl, nil, http.StatusBadRequest)

		// try with empty body
		request(t, "PUT", "/api/v1/access/1", testAccessHandlers.accessUpdateCtrl, nil, http.StatusInternalServerError)

		// try with error in update fn
		request(t, "PUT", "/api/v1/access/1", testAccessHandlers.accessUpdateCtrl, accessData, http.StatusInternalServerError)

	}

}

func Test_accessDeleteCtrl(t *testing.T) {
	testAccessHandlers := accessHandlers{}
	testAccessHandlers.dataStore = prepareAccessMock()

	access := store.Access{
		Name:         "test_access_1",
		Owner:        1,
		Type:         "registry",
		ResourceName: "test_resource",
		Action:       "pull:push",
	}

	// get existed access info by id
	accessData, err := json.Marshal(access)
	require.NoError(t, err)

	// create access first
	request(t, "POST", "/api/v1/access", testAccessHandlers.accessAddCtrl, accessData, http.StatusOK)

	// try to deleteDigest access
	request(t, "DELETE", "/api/v1/access/1", testAccessHandlers.accessDeleteCtrl, accessData, http.StatusOK)

	// try to deleteDigest not existed access
	request(t, "DELETE", "/api/v1/access/1", testAccessHandlers.accessDeleteCtrl, accessData, http.StatusInternalServerError)

	// check for unparsed id
	request(t, "DELETE", "/api/v1/access/abc", testAccessHandlers.accessDeleteCtrl, accessData, http.StatusBadRequest)

}

func prepareAccessMock() engine.Interface {
	testAccessStorage := make(map[int64]store.Access)
	var testAccessIndex int64

	return &engine.InterfaceMock{
		CreateAccessFunc: func(ctx context.Context, access *store.Access) error {

			if _, ok := testAccessStorage[access.ID]; ok {
				return errors.Errorf("access with id [%d] already exist", access.ID)
			}

			testAccessIndex++
			access.ID = testAccessIndex
			testAccessStorage[testAccessIndex] = *access
			return nil
		},
		GetAccessFunc: func(ctx context.Context, id int64) (store.Access, error) {
			if _, ok := testAccessStorage[id]; !ok {
				return store.Access{}, fmt.Errorf("access with id [%d] not found", id)
			}
			return testAccessStorage[id], nil
		},

		FindAccessesFunc: func(ctx context.Context, filter engine.QueryFilter) (engine.ListResponse, error) {
			testListResponse := engine.ListResponse{}

			// fetch by ids
			if filter.Filters != nil {
				if val, ok := filter.Filters["ids"]; ok {

					for _, id := range val.([]interface{}) {
						switch v := id.(type) {
						case float64:
							if val, ok := testAccessStorage[int64(v)]; ok {
								testListResponse.Total++
								testListResponse.Data = append(testListResponse.Data, val)
							}
						}
					}

				}
				if testListResponse.Total == 0 {
					return testListResponse, errors.New("access records not found")
				}
				return testListResponse, nil
			}

			// fetch all records
			for _, v := range testAccessStorage {
				testListResponse.Total++
				testListResponse.Data = append(testListResponse.Data, v)
			}

			return testListResponse, nil
		},

		UpdateAccessFunc: func(ctx context.Context, access store.Access) error {
			// for call error in test only. In real storage access can be to update
			if access.ID == -1 {
				return errors.New("default access can't be updated")
			}

			if _, ok := testAccessStorage[access.ID]; !ok {
				return errors.Errorf("access with id [%d] not found", access.ID)
			}

			testAccessStorage[access.ID] = access
			return nil
		},

		DeleteAccessFunc: func(ctx context.Context, key string, id interface{}) error {
			if _, ok := testAccessStorage[id.(int64)]; !ok {
				return errors.Errorf("access with id [%d] not found", id)
			}
			delete(testAccessStorage, id.(int64))
			return nil
		},
	}
}
