package embedded

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"sync"
	"testing"
	"time"
)

func TestEmbedded_CreateAccess(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	access := &store.Access{
		Name:         "test_access",
		Owner:        9999,
		Type:         "repository",
		ResourceName: "test_rep/test",
		Action:       "delete",
	}

	err := db.CreateAccess(ctx, access)
	require.NoError(t, err)

	// check for duplicates values in fields at existed accesses entries
	err = db.CreateAccess(ctx, access)
	assert.Error(t, err)

	// try with  bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	assert.Error(t, badConn.CreateAccess(ctx, access))

	// check filled required fields
	access = &store.Access{}
	err = db.CreateAccess(ctx, access)
	assert.Error(t, err)
	assert.Equal(t, "required access fields not set: Name, Type, Resource name, Action, Owner", err.Error())

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_GetAccess(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	testAccess := &store.Access{
		Name:         "test_access",
		Owner:        9999,
		Type:         "repository",
		ResourceName: "test_rep/test",
		Action:       "delete",
	}

	err := db.CreateAccess(ctx, testAccess)
	require.NoError(t, err)

	access, err := db.GetAccess(ctx, testAccess.ID)
	require.NoError(t, err)
	assert.Equal(t, access, *testAccess)

	_, err = db.GetAccess(ctx, -1)
	require.Error(t, err)

	// try with  bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	_, err = badConn.GetAccess(ctx, -1)
	assert.Error(t, err)

	ctxCancel()
	wg.Wait()

}

func TestEmbedded_FindAccesses(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	testAccesses := []store.Access{
		{
			Name:         "test_access_1",
			Owner:        1111,
			Type:         "repository",
			ResourceName: "test_rep/test_1",
			Action:       "push",
		},
		{
			Name:         "test_access_2",
			Owner:        2222,
			Type:         "repository",
			ResourceName: "test_rep/test_2",
			Action:       "pull",
		},
		{
			Name:         "test_access_3",
			Owner:        3333,
			Type:         "repository",
			ResourceName: "test_per/test_3",
			Action:       "delete",
		},
		{
			Name:         "test_access_4",
			Owner:        4444,
			IsGroup:      true,
			Type:         "repository",
			ResourceName: "test_per/test_4",
			Action:       "delete",
			Disabled:     true,
		},
	}

	for i, a := range testAccesses {
		tmpAccess := a
		err := db.CreateAccess(ctx, &tmpAccess)
		testAccesses[i].ID = tmpAccess.ID
		require.NoError(t, err)
	}

	accesses, err := db.FindAccesses(ctx, engine.QueryFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(len(accesses.Data)), accesses.Total)
	assert.Equal(t, len(accesses.Data), 4)

	for i, a := range accesses.Data {
		assert.Equal(t, a, testAccesses[i])
	}

	{ // test filter query
		filter := engine.QueryFilter{
			Range:   [2]int64{0, 2},
			Filters: map[string]interface{}{"q": "test_per"},
			Sort:    []string{"id", "asc"},
		}
		accesses, err = db.FindAccesses(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, len(accesses.Data), 2)
	}

	{
		// test filter query with no result
		filter := engine.QueryFilter{
			Range:   [2]int64{0, 2},
			Filters: map[string]interface{}{"q": "unknown_resource"},
			Sort:    []string{"id", "asc"},
		}
		accesses, err = db.FindAccesses(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, int64(0), accesses.Total)
		assert.Equal(t, 0, len(accesses.Data))
	}

	{ // test filter query
		filter := engine.QueryFilter{
			Range:   [2]int64{0, 2},
			Filters: map[string]interface{}{"q": "test_per", "is_group": 1},
			Sort:    []string{"id", "asc"},
		}

		accesses, err = db.FindAccesses(ctx, filter)
		require.NoError(t, err)
		assert.Equal(t, len(accesses.Data), 1)
		assert.Equal(t, accesses.Data[0].(store.Access).Owner, int64(4444))
	}

	// try with  bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	_, err = badConn.FindAccesses(ctx, engine.QueryFilter{})
	assert.Error(t, err)

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_UpdateAccess(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	testAccess := &store.Access{
		Name:         "test_access",
		Owner:        9999,
		Type:         "repository",
		ResourceName: "test_rep/test",
		Action:       "delete",
	}

	err := db.CreateAccess(ctx, testAccess)
	require.NoError(t, err)

	access, err := db.GetAccess(ctx, testAccess.ID)
	require.NoError(t, err)
	assert.Equal(t, access, *testAccess)

	testAccess = &store.Access{
		ID:           testAccess.ID,
		Name:         "update_test_access",
		Owner:        8888,
		Type:         "repository",
		ResourceName: "test_per/test",
		Action:       "pull",
	}

	err = db.UpdateAccess(ctx, *testAccess)
	require.NoError(t, err)

	access, err = db.GetAccess(ctx, testAccess.ID)
	require.NoError(t, err)
	assert.Equal(t, access, *testAccess)

	testAccess.ID = -1
	err = db.UpdateAccess(ctx, *testAccess)
	require.Error(t, err)

	// try with bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	err = badConn.UpdateAccess(ctx, *testAccess)
	assert.Error(t, err)

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_DeleteAccess(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	testAccess := &store.Access{
		Name:         "update_test_access",
		Owner:        8888,
		Type:         "repository",
		ResourceName: "test_per/test",
		Action:       "pull",
	}
	err := db.CreateAccess(ctx, testAccess)
	require.NoError(t, err)

	access, err := db.GetAccess(ctx, testAccess.ID)
	require.NoError(t, err)
	assert.Equal(t, access, *testAccess)

	err = db.DeleteAccess(ctx, "id", testAccess.ID)
	assert.NoError(t, err)

	_, err = db.GetAccess(ctx, testAccess.ID)
	assert.Error(t, err)

	err = db.DeleteAccess(ctx, "id", -1)
	assert.Error(t, err)

	// try with bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	err = badConn.DeleteAccess(ctx, "id", -1)
	assert.Error(t, err)

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_AccessGarbageCollector(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	testAccesses := []store.Access{
		{
			Name:         "test_access_1",
			Owner:        1111,
			Type:         "repository",
			ResourceName: "test_rep/test_1",
			Action:       "push",
		},
		{
			Name:         "test_access_2",
			Owner:        2222,
			Type:         "repository",
			ResourceName: "test_rep/test_2",
			Action:       "pull",
		},
		{
			Name:         "test_access_3",
			Owner:        3333,
			Type:         "repository",
			ResourceName: "test_per/test_3",
			Action:       "delete",
		},
		{
			Name:         "test_access_4",
			Owner:        4444,
			IsGroup:      true,
			Type:         "repository",
			ResourceName: "test_per/test_4",
			Action:       "delete",
			Disabled:     true,
		},
	}
	for i, a := range testAccesses {
		tmpAccess := a
		err := db.CreateAccess(ctx, &tmpAccess)
		testAccesses[i].ID = tmpAccess.ID
		require.NoError(t, err)
	}

	now := time.Now().Unix()
	repositoryEntries := []store.RegistryEntry{
		{
			RepositoryName: "test_rep/test_1",
			Tag:            "test_tag_1",
			Digest:         "sha256:0ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842166",
			Size:           708,
			PullCounter:    1,
			Timestamp:      now,
			Raw:            `{"some":"json_1"}`,
		},
		{
			RepositoryName: "test_rep/test_2",
			Tag:            "test_tag_2",
			Digest:         "sha256:1ea8895f450959fa676bcc1df0611ea93823a735a01205fd8622846041d0c7cf",
			ConfigDigest:   "sha256:2b83bbdc2334fbdb889af0f8e3892255a8b6a32029ffd7fc9e0b3dcd0e842166",
			Size:           709,
			PullCounter:    1,
			Timestamp:      now,
			Raw:            `{"some":"json_2"}`,
		},
	}

	for _, entry := range repositoryEntries {
		tmpGr := entry
		err := db.CreateRepository(ctx, &tmpGr)
		entry.ID = tmpGr.ID
		require.NoError(t, err)
	}

	err := db.AccessGarbageCollector(ctx)
	assert.NoError(t, err)

	filter := engine.QueryFilter{
		Sort: []string{"id", "asc"},
	}

	result, errFind := db.FindRepositories(ctx, filter)
	assert.NoError(t, errFind)
	assert.Equal(t, int64(2), result.Total)

	// try with bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	err = badConn.AccessGarbageCollector(ctx)
	assert.Error(t, err)
	ctxCancel()
	wg.Wait()
}
