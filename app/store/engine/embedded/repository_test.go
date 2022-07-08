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
)

func TestEmbedded_CreateRepository(t *testing.T) {

	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	testEntry := store.RegistryEntry{
		RepositoryName: "hello_test",
		Tag:            "test_tag",
		Digest:         "sha256:fea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
		Size:           708,
		Raw:            []byte(`{"some":"json"}`),
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
		Size:           708,
		Raw:            []byte(`{"some":"json"}`),
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

	entries := []store.RegistryEntry{
		{
			RepositoryName: "aHello_test_1",
			Tag:            "test_tag_1",
			Digest:         "sha256:0ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			Size:           708,
			Raw:            []byte(`{"some":"json_1"}`),
		},
		{
			RepositoryName: "aHello_test_2",
			Tag:            "test_tag_2",
			Digest:         "sha256:1ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			Size:           709,
			Raw:            []byte(`{"some":"json_2"}`),
		},
		{
			RepositoryName: "bHello_test_3",
			Tag:            "test_tag_3",
			Digest:         "sha256:3ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			Size:           710,
			Raw:            []byte(`{"some":"json_3"}`),
		},
		{
			RepositoryName: "bHello_test_4",
			Tag:            "test_tag_4",
			Digest:         "sha256:4ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			Size:           711,
			Raw:            []byte(`{"some":"json_4"}`),
		},
	}

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
	assert.Equal(t, result.Total, int64(2))
	assert.Equal(t, len(result.Data), 2)

	for _, entry := range result.Data {
		u := entry.(store.RegistryEntry)
		if u.RepositoryName != "aHello_test_1" && u.RepositoryName != "aHello_test_2" {
			assert.NoError(t, errors.New("name is expected"))
		}
	}

	// fetch records start with ba* and has disabled field is false
	filter = engine.QueryFilter{
		Range:   [2]int64{0, 2},
		Filters: map[string]interface{}{"repository_name": "aHello_test_1"},
		Sort:    []string{"id", "asc"},
	}

	result, err = db.FindRepositories(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, result.Total, int64(1))
	assert.Equal(t, len(result.Data), 1)

	// fetch with no result
	filter = engine.QueryFilter{
		Range:   [2]int64{0, 2},
		Filters: map[string]interface{}{"repository_name": "unknown_repo_name"},
		Sort:    []string{"id", "asc"},
	}

	result, err = db.FindRepositories(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, result.Total, int64(0))
	assert.Equal(t, len(result.Data), 0)

	// fetch records start with ba* and has disabled field is false
	filter = engine.QueryFilter{
		Range:   [2]int64{0, 5},
		Filters: map[string]interface{}{},
		Sort:    []string{"id", "asc"},
	}

	result, err = db.FindRepositories(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, result.Total, int64(4))
	assert.Equal(t, len(result.Data), 4)

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
