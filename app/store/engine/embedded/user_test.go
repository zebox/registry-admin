package embedded

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"os"
	"sync"
	"testing"
	"time"
)

func TestEmbedded_CreateUser(t *testing.T) {

	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	user := store.User{
		Login:       "test_user",
		Name:        "test_user_name",
		Password:    "test_user_password",
		Role:        "admin",
		Group:       0,
		Disabled:    false,
		Description: "test_description",
	}

	err := db.CreateUser(ctx, &user)
	assert.NoError(t, err)
	assert.NotEqual(t, user.ID, "")
	assert.NotEqual(t, user.Password, "test_user_password")

	{
		badConn := Embedded{}
		err = badConn.Connect(ctx)
		require.NoError(t, err)
		require.NoError(t, badConn.Close(ctx))
		err = badConn.CreateUser(ctx, &user)
		assert.Error(t, err)
	}

	// test with empty required fields
	user2 := &store.User{
		Login:       "",
		Name:        "",
		Password:    "",
		Role:        "unknown",
		Disabled:    false,
		Description: "test_description",
	}
	err = db.CreateUser(ctx, user2)
	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("required user fields not set: Login, Name, Password, role 'unknown' not allowed"))

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_GetUser(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	user := &store.User{
		Login:       "test_user",
		Name:        "test_user_name",
		Password:    "test_user_password",
		Role:        "admin",
		Group:       1,
		Disabled:    false,
		Description: "test_description",
	}

	err := db.CreateUser(ctx, user)
	assert.NoError(t, err)
	assert.NotEqual(t, user.ID, 0)
	assert.NotEqual(t, user.Password, "test_user_password")

	{
		// test with int ID
		userData, err := db.GetUser(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, *user, userData)
	}

	{
		// test with ID as string type
		userData, err := db.GetUser(ctx, "2")
		assert.NoError(t, err)
		assert.Equal(t, *user, userData)
	}
	{
		// test with login as ID
		userData, err := db.GetUser(ctx, "test_user")
		assert.NoError(t, err)
		assert.Equal(t, *user, userData)
	}

	{
		// test with doesn't exist ID
		_, err := db.GetUser(ctx, -1)
		assert.Error(t, err)
	}
	_, err = db.GetUser(ctx, struct{}{})
	assert.Error(t, err)

	// test with connection error
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	_, err = badConn.GetUser(ctx, "test_user")
	assert.Error(t, err)

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_FindUsers(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	users := []store.User{
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
			Group:       1,
			Disabled:    false,
			Description: "baz_description",
		},
		{
			Login:       "qux",
			Name:        "qux",
			Password:    "qux_password",
			Role:        "manager",
			Group:       1,
			Disabled:    false,
			Description: "qux_description",
		},
	}

	for _, user := range users {
		v := user
		err := db.CreateUser(ctx, &v)
		require.NoError(t, err)
	}

	// fetch records start with ba* and has disabled field is false
	filter := engine.QueryFilter{
		Range:   [2]int64{0, 2},
		Filters: map[string]interface{}{"q": "ba", "disabled": 0},
		Sort:    []string{"id", "asc"},
	}

	result, err := db.FindUsers(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, result.Total, int64(2))
	assert.Equal(t, len(result.Data), 2)

	for _, user := range result.Data {
		u := user.(store.User)
		if u.Name != "baz" && u.Name != "bar" {
			assert.NoError(t, errors.New("name is expected"))
		}
	}

	// with empty search string and  strong field condition
	filter.Filters = map[string]interface{}{"q": "", "disabled": 1, "login": "foo"}
	result, err = db.FindUsers(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, result.Total, int64(1))
	require.Equal(t, len(result.Data), 1)

	u := result.Data[0].(store.User)
	assert.Equal(t, "foo", u.Name)

	// fetch all records
	filter.Filters = map[string]interface{}{"q": ""}
	filter.Range = [2]int64{0, 0} // reset range

	result, err = db.FindUsers(ctx, filter)
	assert.NoError(t, err)

	// total is 5 (five) because default user added when DB init
	assert.Equal(t, result.Total, int64(5))
	require.Equal(t, len(result.Data), 5)

	// test bas query syntax
	filter.Filters = map[string]interface{}{"q": "", "role LIKE": "%%%%"}
	filter.Range = [2]int64{0, 0} // reset range

	result, err = db.FindUsers(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Data))

	// test for sql-injection
	filter.Filters = map[string]interface{}{"q": "", "role LIKE '%%') --": "user"}
	filter.Range = [2]int64{0, 0} // reset range

	result, err = db.FindUsers(ctx, filter)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result.Data))

	// test with connection error
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	_, err = badConn.FindUsers(ctx, filter)
	assert.Error(t, err)

	assert.NoError(t, db.db.Close())
	ctxCancel()
	wg.Wait()
}

func TestEmbedded_UpdateUser(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	user := &store.User{
		Login:       "test_user",
		Name:        "test_user_name",
		Password:    "test_user_password",
		Role:        "admin",
		Group:       1,
		Disabled:    false,
		Description: "test_description",
	}

	err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	user.Name = "new_user_name"
	user.Password = "new_user_password"
	user.Role = "unknown"
	assert.Error(t, db.UpdateUser(ctx, *user))

	user.Role = "manager"
	assert.NoError(t, db.UpdateUser(ctx, *user))
	assert.NoError(t, user.HashAndSalt()) // hash password for compare

	userData, err := db.GetUser(ctx, user.ID)
	assert.NoError(t, err)
	assert.Equal(t, userData.Name, user.Name)
	assert.Equal(t, userData.Role, user.Role)
	assert.True(t, store.ComparePassword(user.Password, "new_user_password"))

	// try update without password change
	user.Name = "updated_user_name"
	user.Password = ""
	assert.NoError(t, db.UpdateUser(ctx, *user))
	userData, err = db.GetUser(ctx, user.ID)
	require.NoError(t, err)
	assert.True(t, store.ComparePassword(userData.Password, "new_user_password"))
	assert.Equal(t, "updated_user_name", userData.Name)

	// try to update not existed user
	user.ID = -1
	assert.Error(t, db.UpdateUser(ctx, *user))

	// test with connection error
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	assert.Error(t, badConn.UpdateUser(ctx, *user))

	ctxCancel()
	wg.Wait()
}

func TestEmbedded_DeleteUser(t *testing.T) {
	ctx, ctxCancel := context.WithCancel(context.Background())
	var wg = new(sync.WaitGroup)
	db := prepareTestDB(ctx, t, wg) // defined mock store

	user := &store.User{
		Login:       "test_user",
		Name:        "test_user_name",
		Password:    "test_user_password",
		Role:        "admin",
		Group:       1,
		Disabled:    false,
		Description: "test_description",
	}

	err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	assert.Error(t, db.DeleteUser(ctx, 9999))
	assert.NoError(t, db.DeleteUser(ctx, user.ID))
	_, err = db.GetUser(ctx, user.ID)
	assert.Error(t, err)

	// test with connection error
	badConn := Embedded{}
	err = badConn.Connect(ctx)
	require.NoError(t, err)
	require.NoError(t, badConn.Close(ctx))
	assert.Error(t, badConn.DeleteUser(ctx, -1))

	ctxCancel()
	wg.Wait()
}

func prepareTestDB(ctx context.Context, t *testing.T, wg *sync.WaitGroup) *Embedded {
	testDBPath := os.TempDir() + "/test.db"

	_ = os.Remove(testDBPath)

	sqlite := NewEmbedded(testDBPath)
	err := sqlite.Connect(ctx)
	require.NoError(t, err)
	wg.Add(1)
	go func() {

		<-ctx.Done()
		err = sqlite.Close(ctx)
		assert.NoError(t, err)
		time.Sleep(time.Millisecond * 50) // wait for close connection
		err = os.Remove(testDBPath)
		assert.NoError(t, err)
		wg.Done()
	}()

	return sqlite
}
