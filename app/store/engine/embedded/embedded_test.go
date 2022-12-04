package embedded

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store/engine"
	"os"
	"testing"
)

func TestSQLite_Connect(t *testing.T) {
	dbPath := os.TempDir() + "/test.db"
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()

	_ = os.Remove(dbPath) // clear if exist on previously tests
	db := Embedded{Path: dbPath}

	err := db.Connect(ctx)
	require.NoError(t, err)
	assert.NotNil(t, db.db)

	isExist, err := db.isTableExist(ctx, usersTable)
	assert.NoError(t, err)
	assert.True(t, isExist)

	isExist, err = db.isTableExist(ctx, groupsTable)
	assert.NoError(t, err)
	assert.True(t, isExist)

	isExist, err = db.isTableExist(ctx, accessTable)
	assert.NoError(t, err)
	assert.True(t, isExist)

	isExist, err = db.isTableExist(ctx, repositoriesTable)
	assert.NoError(t, err)
	assert.True(t, isExist)

	assert.NoError(t, db.Close(ctx))
	assert.NoError(t, os.Remove(dbPath))

	t.Log("test with bad db path ")
	dbPath = os.TempDir() + "/unknown_path/test.db"
	db = Embedded{Path: dbPath}
	err = db.Connect(ctx)
	require.Error(t, err)
}

func TestSQlite_initTables(t *testing.T) {
	dbPath := os.TempDir() + "/test.db"
	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	db := Embedded{Path: dbPath}

	var err error
	db.db, err = sql.Open("sqlite3", db.Path)
	require.NoError(t, err)

	assert.NoError(t, db.initUserTable(ctx))
	assert.NoError(t, db.initGroupsTable(ctx))
	assert.NoError(t, db.initAccessTable(ctx))
	assert.NoError(t, db.initRepositoriesTable(ctx))

	err = db.initTables(ctx)
	assert.Error(t, err)

	assert.NoError(t, db.Close(ctx))
	_ = os.Remove(dbPath)

}
func TestNewEmbedded(t *testing.T) {
	testPathToDB := "/var/test/store.db"
	embedded := NewEmbedded(testPathToDB)
	assert.Equal(t, embedded.Path, testPathToDB)
}

func TestFilterBuilder(t *testing.T) {

	// http://192.168.12.58/api/v1/sessions?filter={"status":2000,"area_id":3276989022,"q":"пр"}&range=[0,24]&sort=["start_time","DESC"]
	filter := engine.QueryFilter{
		Range:   [2]int64{1, 10},
		Filters: map[string]interface{}{"q": "test", "disabled": 1},
		Sort:    []string{"id", "asc"},
	}

	{
		f := filtersBuilder(filter, "role", "login")
		checkWhere := "WHERE (role LIKE '%test%' OR login LIKE '%test%') AND (disabled = 1) ORDER BY id asc  LIMIT 9 OFFSET 1"
		assert.Equal(t, checkWhere, f.allClauses)
	}

	{
		filter.Filters = map[string]interface{}{"q": "test"}
		f1 := filtersBuilder(filter, "role", "login")
		checkWhere := "WHERE (role LIKE '%test%' OR login LIKE '%test%') ORDER BY id asc  LIMIT 9 OFFSET 1"
		assert.Equal(t, checkWhere, f1.allClauses)
	}
	{

		ids := []interface{}{float64(1019101756), float64(1334517373)}
		filter.Filters = map[string]interface{}{"q": "test", "disabled": 1}
		filter.Filters["ids"] = ids
		f2 := filtersBuilder(filter, "role", "login")
		checkWhere2 := "WHERE id IN (1019101756, 1334517373) AND (role LIKE '%test%' OR login LIKE '%test%') AND (disabled = 1) ORDER BY id asc  LIMIT 9 OFFSET 1"
		assert.Equal(t, checkWhere2, f2.allClauses)
	}
	{
		delete(filter.Filters, "ids")
		filter.Range = [2]int64{}
		filter.Filters = map[string]interface{}{"q": 1}
		f2 := filtersBuilder(filter, "role", "id")
		checkWhere2 := "WHERE ( role LIKE 1 OR  id LIKE 1) ORDER BY id asc "
		assert.Equal(t, checkWhere2, f2.allClauses)
	}

	{
		// test for sanitize key/value
		filter.Filters = map[string]interface{}{"q select--": "test query WHERE LIKE JOIN search DELETE -- % = string ", "description": "-- LIKE AND SELECT clear_value WHERE OR"}
		f := filtersBuilder(filter, "role", "login")
		checkWhere := "WHERE (role LIKE '%test querysearchstring %' OR login LIKE '%test querysearchstring %') AND (description = 'ANDclear_valueOR') ORDER BY id asc "
		assert.Equal(t, checkWhere, f.allClauses)
	}

}
