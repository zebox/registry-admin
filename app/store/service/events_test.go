package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"github.com/docker/distribution/notifications"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"strings"
	"testing"
	"time"
)

func TestDataService_RepositoryEventsProcessing(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ds := DataService{
		Storage: prepareEngineMock(),
	}
	testEnvelope := notifications.Envelope{
		Events: []notifications.Event{
			{
				ID:        "320678d8-ca14-430f-8bb6-4ca139cd83f7",
				Timestamp: time.Now(),
				Action:    notifications.EventActionPush,
				Request: notifications.RequestRecord{
					ID:        "6df24a34-0959-4923-81ca-14f09767db19",
					Addr:      "192.168.64.11:42961",
					Host:      "192.168.100.227:5000",
					Method:    "GET",
					UserAgent: "curl/7.38.0",
				},
				Actor: notifications.ActorRecord{},
				Source: notifications.SourceRecord{
					Addr:       "xtal.local:5000",
					InstanceID: "a53db899-3b4b-4a62-a067-8dd013beaca4",
				},
			},
		},
	}

	createTestEvent := func() {
		testEnvelope.Events[0].Target.Repository = "test/repo_1"
		testEnvelope.Events[0].Target.Tag = "1.1.0"
		testEnvelope.Events[0].Target.MediaType = "application/vnd.docker.distribution.manifest.v2+json"
		testEnvelope.Events[0].Target.Digest = "sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf"
		testEnvelope.Events[0].Target.Length = 708
		testEnvelope.Events[0].Target.Size = 708
		testEnvelope.Events[0].Target.URL = "http://192.168.100.227:5000/v2/hello-world/manifests/sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf"
	}

	createTestEvent()

	err := ds.RepositoryEventsProcessing(ctx, testEnvelope)
	assert.NoError(t, err)

	// test with pull action event
	testEnvelopePullEvent := testEnvelope
	testEnvelopePullEvent.Events[0].Action = notifications.EventActionPull
	for i := 0; i < 3; i++ {
		errProcessing := ds.RepositoryEventsProcessing(ctx, testEnvelopePullEvent)
		assert.NoError(t, errProcessing)
	}
	filter := engine.QueryFilter{
		Filters: map[string]interface{}{"repository_name": testEnvelopePullEvent.Events[0].Target.Repository, "tag": testEnvelopePullEvent.Events[0].Target.Tag},
	}

	result, errFind := ds.Storage.FindRepositories(ctx, filter)
	assert.NoError(t, errFind)
	require.Len(t, result.Data, 1)
	assert.IsType(t, store.RegistryEntry{}, result.Data[0])
	testRegistyEntry := result.Data[0].(store.RegistryEntry)
	assert.Equal(t, int64(3), testRegistyEntry.PullCounter)

	// test with not exist repository
	testEnvelope.Events[0].Target.Repository = "test/repo_3"
	testEnvelope.Events[0].Target.Tag = "1.1.0"
	err = ds.RepositoryEventsProcessing(ctx, testEnvelope)
	assert.NoError(t, err)

	// test with multiple values
	testEnvelope.Events[0].Target.Repository = "test/repo_"
	testEnvelope.Events[0].Target.Tag = "1."
	err = ds.RepositoryEventsProcessing(ctx, testEnvelope)
	assert.Error(t, err)

	// test with filter error
	testEnvelope.Events[0].Target.Repository = ""
	testEnvelope.Events[0].Target.Tag = ""
	err = ds.RepositoryEventsProcessing(ctx, testEnvelope)
	assert.Error(t, err)

	// test with delete action
	createTestEvent()
	testEnvelopePullEvent.Events[0].Action = notifications.EventActionDelete
	err = ds.RepositoryEventsProcessing(ctx, testEnvelope)
	assert.NoError(t, err)

	// test delete with not existed repository entry
	testEnvelopePullEvent.Events[0].Target.Repository = "unknown"
	err = ds.RepositoryEventsProcessing(ctx, testEnvelope)
	assert.Error(t, err)

	// test delete with not existed repository entry
	testEnvelopePullEvent.Events[0].Target.Repository = "unknown"
	err = ds.RepositoryEventsProcessing(nil, testEnvelope) // nolint
	assert.Error(t, err)
}

func prepareEngineMock() *engine.InterfaceMock {

	testRepositoriesEntries := []store.RegistryEntry{
		{
			ID:             1,
			RepositoryName: "test/repo_1",
			Tag:            "1.1.0",
		},
		{
			ID:             2,
			RepositoryName: "test/repo_1",
			Tag:            "1.2.0",
		},
		{
			ID:             3,
			RepositoryName: "test/repo_2",
			Tag:            "2.1.0",
		},
		{
			ID:             4,
			RepositoryName: "test/repo_2",
			Tag:            "2.2.0",
		},
	}

	return &engine.InterfaceMock{
		CreateRepositoryFunc: func(ctx context.Context, entry *store.RegistryEntry) error {

			for _, testEntry := range testRepositoriesEntries {
				if testEntry.RepositoryName == entry.RepositoryName && testEntry.Tag == entry.Tag {
					return errors.New("duplicate repository entry")
				}

				hasher := sha256.New()
				hasher.Write([]byte(entry.RepositoryName + entry.Tag))
				sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
				entry.Digest = "sha256:" + sha
				entry.Timestamp = time.Now().Unix()

			}
			return nil
		},

		UpdateRepositoryFunc: func(ctx context.Context, conditionClause map[string]interface{}, data map[string]interface{}) error {
			var id int64
			if v, ok := conditionClause["id"]; ok {
				id = v.(int64)
			}
			for i, testEntry := range testRepositoriesEntries {
				if id == testEntry.ID {
					for k, v := range data {
						if k == "pull_counter" {
							testRepositoriesEntries[i].PullCounter = v.(int64)
						}
					}
					return nil
				}
			}
			return errors.New("entry not found")
		},

		FindRepositoriesFunc: func(ctx context.Context, filter engine.QueryFilter) (result engine.ListResponse, err error) {

			if ctx == nil {
				return result, errors.New("nil context not allowed")
			}
			if _, ok := filter.Filters["repository_name"]; !ok {
				return result, errors.New("empty repository name not allowed")
			}
			searchRepo := filter.Filters["repository_name"].(string)

			if _, ok := filter.Filters["tag"]; !ok {
				return result, errors.New("empty repository name not allowed")
			}
			searchTag := filter.Filters["tag"].(string)

			if searchRepo == "" || searchTag == "" {
				return result, errors.New("empty repository name or tag not allowed")
			}
			var counter int64
			for _, rep := range testRepositoriesEntries {

				if strings.HasPrefix(rep.RepositoryName, searchRepo) && strings.HasPrefix(rep.Tag, searchTag) {
					counter++
					result = engine.ListResponse{
						Total: counter,
						Data:  append(result.Data, rep),
					}
				}
			}
			return result, nil
		},

		DeleteRepositoryFunc: func(ctx context.Context, key string, id interface{}) error {
			for _, val := range testRepositoriesEntries {
				if val.ID == id.(int64) {
					return nil
				}
			}
			return errors.Errorf("entry not found: id %v", id)
		},
	}
}
