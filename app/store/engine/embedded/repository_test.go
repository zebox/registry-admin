package embedded

import (
	"context"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"sync"
	"testing"
	"time"
)

var entries = []store.RegistryEntry{
	{
		RepositoryName: "aHello_test_1",
		Tag:            "test_tag_1",
		Digest:         "sha256:0ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7c1",
		ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842166",
		Size:           708,
		PullCounter:    1,
		Timestamp:      time.Now().Unix(),
		Raw:            `{"some":"json_1"}`,
	},
	{
		RepositoryName: "aHello_test_1",
		Tag:            "test_tag_1_2",
		Digest:         "sha256:0ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7bb",
		ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842166",
		Size:           1500,
		PullCounter:    1,
		Timestamp:      time.Now().Unix(),
		Raw:            `{"some":"json_2"}`,
	},
	{
		RepositoryName: "aHello_test_2",
		Tag:            "test_tag_2",
		Digest:         "sha256:1ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7c2",
		ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842165",
		Size:           709,
		PullCounter:    1,
		Timestamp:      time.Now().Unix(),
		Raw:            `{"some":"json_2"}`,
	},
	{
		RepositoryName: "bHello_test_3",
		Tag:            "test_tag_3",
		Digest:         "sha256:3ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7c3",
		ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842164",
		Size:           710,
		PullCounter:    1,
		Timestamp:      time.Now().Unix(),
		Raw:            `{"some":"json_3"}`,
	},
	{
		RepositoryName: "bHello_test_4",
		Tag:            "test_tag_4",
		Digest:         "sha256:4ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7c4",
		ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842163",
		Size:           711,
		PullCounter:    1,
		Timestamp:      time.Now().Unix(),
		Raw:            `{"some":"json_4"}`,
	},
	{
		RepositoryName: "bHello_test_4",
		Tag:            "test_tag_4_1",
		Digest:         "sha256:4ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7c5",
		ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842162",
		Size:           711,
		PullCounter:    1,
		Timestamp:      time.Now().Unix(),
		Raw:            `{"some":"json_4_1"}`,
	},
	{
		RepositoryName: "bHello_test_4",
		Tag:            "test_tag_4_2",
		Digest:         "sha256:4ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7c6",
		ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842161",
		Size:           711,
		PullCounter:    1,
		Timestamp:      time.Now().Unix(),
		Raw:            `{"some":"json_4_2"}`,
	},
}

func TestEmbedded_CreateRepository(t *testing.T) {

	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store
	testEntry := store.RegistryEntry{
		RepositoryName: "hello_test",
		Tag:            "test_tag",
		Digest:         "sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
		ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842162",
		Size:           708,
		PullCounter:    1,
		Timestamp:      time.Now().Unix(),
		Raw:            `{"some":"json"}`,
	}

	err := db.CreateRepository(ctx, &testEntry)
	assert.NoError(t, err)
	assert.Greater(t, testEntry.ID, int64(0))

	// test for duplicate name entry
	err = db.CreateRepository(ctx, &testEntry)
	require.NotNil(t, err)
	assert.Equal(t, err.Error(), "UNIQUE constraint failed: repositories.repository_name, repositories.tag")

	// test with empty required fields
	testEntry.RepositoryName = ""
	err = db.CreateRepository(ctx, &testEntry)
	require.NotNil(t, err)
	assert.Equal(t, err.Error(), "CHECK constraint failed: repository_name <> ''")

	// try with  bad or closed connection
	testEntry.RepositoryName = "test_new_group"
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	assert.Error(t, badConn.CreateRepository(ctx, &testEntry))

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_GetRepository(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	testEntry := store.RegistryEntry{
		RepositoryName: "hello_test",
		Tag:            "test_tag",
		Digest:         "sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
		ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842162",
		Size:           708,
		PullCounter:    1,
		Timestamp:      time.Now().Unix(),
		Raw:            `{"some":"json"}`,
	}
	err := db.CreateRepository(ctx, &testEntry)
	assert.NoError(t, err)
	assert.Greater(t, testEntry.ID, int64(0))

	var e store.RegistryEntry
	e, err = db.GetRepository(ctx, testEntry.ID)
	require.NoError(t, err)
	assert.Equal(t, testEntry, e)

	// test with try to get testEntry with not existed id
	_, err = db.GetRepository(ctx, -1)
	require.Error(t, err)

	// try with  bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	_, err = badConn.GetRepository(ctx, -1)
	assert.Error(t, err)

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_FindRepositories(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	for _, entry := range entries {
		tmpGr := entry
		err := db.CreateRepository(ctx, &tmpGr)
		require.NoError(t, err)
	}

	// fetch records start with ba* and has disabled field is false
	filter := engine.QueryFilter{
		Range:   [2]int64{0, 2},
		Filters: map[string]interface{}{"q": "aHello_test"},
		Sort:    []string{"id", "asc"},
	}

	result, err := db.FindRepositories(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), result.Total)
	assert.Equal(t, 2, len(result.Data))

	for _, entry := range result.Data {
		u := entry.(store.RegistryEntry)
		if u.RepositoryName != "aHello_test_1" && u.RepositoryName != "aHello_test_2" {
			assert.NoError(t, errors.New("name is expected"))
		}
	}

	// fetch records start with ba* and has disabled field is false
	filter = engine.QueryFilter{
		Range:   [2]int64{0, 2},
		Filters: map[string]interface{}{store.RegistryRepositoryNameField: "aHello_test_1"},
		Sort:    []string{"id", "asc"},
	}

	result, err = db.FindRepositories(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.Total)
	assert.Equal(t, 2, len(result.Data))

	// fetch with no result
	filter = engine.QueryFilter{
		Range:   [2]int64{0, 2},
		Filters: map[string]interface{}{store.RegistryRepositoryNameField: "unknown_repo_name"},
		Sort:    []string{"id", "asc"},
	}

	result, err = db.FindRepositories(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), result.Total)
	assert.Equal(t, 0, len(result.Data))

	// fetch records start with ba* and has disabled field is false
	filter = engine.QueryFilter{
		Range:   [2]int64{0, 5},
		Filters: map[string]interface{}{},
		Sort:    []string{"id", "asc"},
	}

	result, err = db.FindRepositories(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(entries)), result.Total)
	assert.Equal(t, 5, len(result.Data)) // used a range limit filter value, total shouldn't equal to result

	// test with 'Distinct' filter value
	filter.Range = [2]int64{}
	filter.GroupByField = true
	result, err = db.FindRepositories(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(4), result.Total)
	assert.Equal(t, 4, len(result.Data))

	// test for size summary when 'Distinct' filter using
	filter.Filters = map[string]interface{}{store.RegistryRepositoryNameField: "aHello_test_1"}
	filter.GroupByField = true
	result, err = db.FindRepositories(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.Total)
	require.Equal(t, 1, len(result.Data))

	reposSize := entries[0].Size + entries[1].Size
	assert.Equal(t, reposSize, result.Data[0].(store.RegistryEntry).Size)

	// try with  bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	_, err = badConn.FindRepositories(ctx, engine.QueryFilter{})
	assert.Error(t, err)

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_FindRepositoriesByUser(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	testUser := store.User{
		Login:    "test_user",
		Name:     "test_user",
		Group:    1,
		Password: "test_password",
		Role:     store.UserRole,
	}

	err := db.CreateUser(ctx, &testUser)
	require.NoError(t, err)

	testAccesses := []store.Access{
		{
			Name:         "test_access_1",
			Owner:        testUser.ID,
			ResourceName: "aHello_test_1",
			Type:         "registry",
			Action:       "pull",
		},
		{
			Name:         "test_access_2",
			Owner:        testUser.ID,
			Type:         "repository",
			ResourceName: "aHello_test_2",
			Action:       "push",
		},
	}

	for _, entry := range entries {
		tmpEntry := entry
		err = db.CreateRepository(ctx, &tmpEntry)
		require.NoError(t, err)
	}

	for _, access := range testAccesses {
		tmpAccess := access
		err = db.CreateAccess(ctx, &tmpAccess)
		require.NoError(t, err)
	}

	// fetch records start with ba* and has disabled field is false
	filter := engine.QueryFilter{
		Filters: map[string]interface{}{engine.RepositoriesByUserAccess: testUser.ID},
	}

	result, errFind := db.FindRepositories(ctx, filter)
	assert.NoError(t, errFind)
	assert.Equal(t, int64(3), result.Total)
	assert.Equal(t, 3, len(result.Data))

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_UpdateRepository(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	for _, entry := range entries {
		tmpGr := entry
		err := db.CreateRepository(ctx, &tmpGr)
		entry.ID = tmpGr.ID
		require.NoError(t, err)
	}

	// test for update a one filed with on conditions
	conditionClause := map[string]interface{}{store.RegistryRepositoryNameField: "aHello_test_2"}
	fieldForUpdate := map[string]interface{}{store.RegistryTagField: "test_tag_222"}
	err := db.UpdateRepository(ctx, conditionClause, fieldForUpdate)
	require.NoError(t, err)

	updatedEntry, errGet := db.GetRepository(ctx, 3)
	require.NoError(t, errGet)
	assert.Equal(t, "test_tag_222", updatedEntry.Tag)

	timestamp := updatedEntry.Timestamp + 100
	conditionClause = map[string]interface{}{store.RegistryRepositoryNameField: "aHello_test_2", store.RegistrySizeNameField: 709}
	fieldForUpdate = map[string]interface{}{store.RegistryTagField: "test_tag_0222", store.RegistryTimestampField: timestamp}
	err = db.UpdateRepository(ctx, conditionClause, fieldForUpdate)
	require.NoError(t, err)

	updatedEntry, errGet = db.GetRepository(ctx, 3)
	require.NoError(t, errGet)
	assert.Equal(t, "test_tag_0222", updatedEntry.Tag)
	assert.Equal(t, timestamp, updatedEntry.Timestamp)

	// try to update not existed repository
	conditionClause = map[string]interface{}{store.RegistryRepositoryNameField: "xyz"}
	fieldForUpdate = map[string]interface{}{store.RegistryTagField: "test_tag_000"}
	assert.Error(t, db.UpdateRepository(ctx, conditionClause, fieldForUpdate))

	// try with  bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	assert.Error(t, badConn.UpdateRepository(ctx, conditionClause, fieldForUpdate))

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_DeleteRepository(t *testing.T) {

	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	testEntry := store.RegistryEntry{
		RepositoryName: "hello_test",
		Tag:            "test_tag",
		Digest:         "sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
		ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842162",
		Size:           708,
		PullCounter:    1,
		Timestamp:      time.Now().Unix(),
		Raw:            `{"some":"json"}`,
	}
	err := db.CreateRepository(ctx, &testEntry)
	assert.NoError(t, err)
	assert.Greater(t, testEntry.ID, int64(0))

	err = db.DeleteRepository(ctx, testEntry.RepositoryName, testEntry.Digest)
	assert.NoError(t, err)

	_, err = db.GetRepository(ctx, testEntry.ID)
	require.Error(t, err)

	err = db.DeleteRepository(ctx, testEntry.RepositoryName, testEntry.Digest)
	assert.Equal(t, err, engine.ErrNotFound)

	err = db.DeleteRepository(ctx, "invalid_name", "unknown")
	assert.Error(t, err)

	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	assert.Error(t, badConn.DeleteRepository(ctx, store.RegistryRepositoryNameField, "hello_test"))

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_RepositoryGarbageCollector(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	dateSync := time.Now().Unix()
	outdated := time.Now().Add(-1 * time.Hour).Unix()
	testEntries := []store.RegistryEntry{
		{
			RepositoryName: "aHello_test_1",
			Tag:            "test_tag_1",
			Digest:         "sha256:0ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842166",
			Size:           708,
			PullCounter:    1,
			Timestamp:      dateSync,
			Raw:            `{"some":"json_1"}`,
		},
		{
			RepositoryName: "aHello_test_2",
			Tag:            "test_tag_2",
			Digest:         "sha256:1ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842116",
			Size:           709,
			PullCounter:    1,
			Timestamp:      dateSync,
			Raw:            `{"some":"json_2"}`,
		},
		{
			RepositoryName: "bHello_test_3",
			Tag:            "test_tag_3",
			Digest:         "sha256:3ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842163",
			Size:           710,
			PullCounter:    1,
			Timestamp:      outdated,
			Raw:            `{"some":"json_3"}`,
		},
		{
			RepositoryName: "bHello_test_4",
			Tag:            "test_tag_4",
			Digest:         "sha256:4ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e84216d",
			Size:           711,
			PullCounter:    1,
			Timestamp:      outdated,
			Raw:            `{"some":"json_4"}`,
		},
	}

	for _, entry := range testEntries {
		tmpGr := entry
		err := db.CreateRepository(ctx, &tmpGr)
		entry.ID = tmpGr.ID
		require.NoError(t, err)
	}

	err := db.RepositoryGarbageCollector(ctx, dateSync)
	assert.NoError(t, err)

	filter := engine.QueryFilter{
		Sort: []string{"id", "asc"},
	}

	result, errFind := db.FindRepositories(ctx, filter)
	assert.NoError(t, errFind)
	assert.Equal(t, int64(2), result.Total)

	// try with  bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	assert.Error(t, badConn.RepositoryGarbageCollector(ctx, 0))

	ctxCancel()
	wg.Wait()
}
