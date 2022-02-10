package embedded

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"sync"
	"testing"
)

func TestEmbedded_CreateGroup(t *testing.T) {

	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	group := store.Group{
		Name:        "test_group_name",
		Description: "test_group_description",
	}

	err := db.CreateGroup(ctx, &group)
	assert.NoError(t, err)
	assert.Greater(t, group.ID, int64(0))

	// test for duplicate name entry
	err = db.CreateGroup(ctx, &group)
	require.NotNil(t, err)
	assert.Equal(t, err.Error(), "UNIQUE constraint failed: groups.name")

	// test with empty required fields
	group.Name = ""
	err = db.CreateGroup(ctx, &group)
	assert.Equal(t, err, ErrRequiredFieldInGroupIsEmpty)

	// try with  bad or closed connection
	group.Name = "test_new_group"
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	assert.Error(t, badConn.CreateGroup(ctx, &group))

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_GetGroup(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	group := store.Group{
		Name:        "test_group_name",
		Description: "test_group_description",
	}

	err := db.CreateGroup(ctx, &group)
	assert.NoError(t, err)
	assert.Greater(t, group.ID, int64(0))

	var g store.Group
	g, err = db.GetGroup(ctx, group.ID)
	require.NoError(t, err)
	assert.Equal(t, g.Name, "test_group_name")
	assert.Equal(t, g.Description, "test_group_description")

	// test with try to get group with not existed id
	g, err = db.GetGroup(ctx, -1)
	require.Error(t, err)

	// try with  bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	_, err = badConn.GetGroup(ctx, -1)
	assert.Error(t, err)

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_FindGroups(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	groups := []store.Group{
		{
			Name:        "foo",
			Description: "foo_description",
		},
		{
			Name:        "bar",
			Description: "bar_description",
		},
		{
			Name:        "baz",
			Description: "baz_description",
		},
		{
			Name:        "qux",
			Description: "qux_description",
		},
	}

	for _, group := range groups {
		tmpGr := group
		err := db.CreateGroup(ctx, &tmpGr)
		require.NoError(t, err)
	}

	// fetch records start with ba* and has disabled field is false
	filter := engine.QueryFilter{
		Range:   [2]int64{0, 2},
		Filters: map[string]interface{}{"q": "ba"},
		Sort:    []string{"id", "asc"},
	}

	result, err := db.FindGroups(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, result.Total, int64(2))
	assert.Equal(t, len(result.Data), 2)

	for _, group := range result.Data {
		u := group.(store.Group)
		if u.Name != "baz" && u.Name != "bar" {
			assert.NoError(t, errors.New("name is expected"))
		}
	}

	// fetch records start with ba* and has disabled field is false
	filter = engine.QueryFilter{
		Range:   [2]int64{0, 2},
		Filters: map[string]interface{}{"name": "qux"},
		Sort:    []string{"id", "asc"},
	}

	result, err = db.FindGroups(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, result.Total, int64(1))
	assert.Equal(t, len(result.Data), 1)

	// fetch with no result
	filter = engine.QueryFilter{
		Range:   [2]int64{0, 2},
		Filters: map[string]interface{}{"name": "fux"},
		Sort:    []string{"id", "asc"},
	}

	result, err = db.FindGroups(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, result.Total, int64(0))
	assert.Equal(t, len(result.Data), 0)

	// fetch records start with ba* and has disabled field is false
	filter = engine.QueryFilter{
		Range:   [2]int64{0, 5},
		Filters: map[string]interface{}{},
		Sort:    []string{"id", "asc"},
	}

	result, err = db.FindGroups(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, result.Total, int64(5))
	assert.Equal(t, len(result.Data), 5)

	// try with  bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	_, err = badConn.FindGroups(ctx, engine.QueryFilter{})
	assert.Error(t, err)

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_UpdateGroup(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	group := &store.Group{

		Name:        "test_group_name",
		Description: "test_description",
	}

	err := db.CreateGroup(ctx, group)
	require.NoError(t, err)

	group.Name = "new_group_name"
	group.Description = "new group description"
	require.NoError(t, db.UpdateGroup(ctx, *group))

	groupData, err := db.GetGroup(ctx, group.ID)
	assert.NoError(t, err)
	assert.Equal(t, groupData.Name, group.Name)
	assert.Equal(t, groupData.Description, group.Description)

	// try to update not existed group
	group.ID = -1
	assert.Error(t, db.UpdateGroup(ctx, *group))

	// try with  bad or closed connection
	group.Name = "test_new_group"
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	assert.Error(t, badConn.UpdateGroup(ctx, *group))

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_DeleteGroup(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	group := &store.Group{
		Name:        "test_user_name",
		Description: "test_description",
	}

	err := db.CreateGroup(ctx, group)
	require.NoError(t, err)

	assert.Error(t, db.DeleteGroup(ctx, -1))
	assert.NoError(t, db.DeleteGroup(ctx, group.ID))
	_, err = db.GetGroup(ctx, group.ID)
	assert.Error(t, err)

	// try with  bad or closed connection
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	assert.Error(t, badConn.DeleteUser(ctx, -1))

	ctxCancel()
	wg.Wait()
}
