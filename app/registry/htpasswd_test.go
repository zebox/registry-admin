package registry

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
	"testing"
)

func TestRegistry_UpdateHtpasswd(t *testing.T) {
	var testUsers = make([]store.User, 0)

	// filling test users store
	for i := 0; i < 10; i++ {
		user := store.User{
			Login:    fmt.Sprintf("user_%d", i),
			Password: fmt.Sprintf("password_%d", i),
		}
		require.NoError(t, user.HashAndSalt())
		testUsers = append(testUsers, user)
	}

	tra := newTestUsersRegistryAdapter(context.Background(), engine.QueryFilter{})
	testPath := os.TempDir() + "/test/.htpasswd"

	r := Registry{htpasswd: &htpasswd{path: testPath}}
	err := r.htpasswd.update(testUsers)
	require.NoError(t, err)

	require.NoError(t, r.UpdateHtpasswd(tra))

	defer func() {
		assert.NoError(t, os.RemoveAll(os.TempDir()+"/test/"))
	}()

	entries := htpasswdReader(t, testPath)
	assert.Equal(t, 10, len(entries))

	for k, v := range entries {
		keySuffix := k[len(k)-2:]
		err := bcrypt.CompareHashAndPassword(v, []byte("password"+keySuffix))
		assert.NoError(t, err)
	}
	assert.NoError(t, r.UpdateHtpasswd(tra))

	r.htpasswd.path = ""
	assert.Error(t, r.UpdateHtpasswd(tra))

	// test for error with nil userFn
	assert.Error(t, r.UpdateHtpasswd(nil))

	r.htpasswd = nil
	assert.Nil(t, r.UpdateHtpasswd(tra))

}

func htpasswdReader(t *testing.T, path string) map[string][]byte {
	entries := map[string][]byte{}
	f, err := os.Open(path)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, f.Close())
	}()

	scanner := bufio.NewScanner(f)
	var line int
	for scanner.Scan() {
		line++ // 1-based line numbering
		text := strings.TrimSpace(scanner.Text())

		if len(text) < 1 {
			continue
		}

		// lines that *begin* with a '#' are considered comments
		if text[0] == '#' {
			continue
		}

		i := strings.Index(text, ":")
		if i < 0 || i >= len(text) {
			require.FailNow(t, "htpasswd: invalid entry at line %d: %q", line, scanner.Text())
		}

		entries[text[:i]] = []byte(text[i+1:])
	}

	if err := scanner.Err(); err != nil {
		require.FailNow(t, "htpasswd: invalid entry at line %v", err)
	}
	return entries
}

// uses for bind FindUsers func in store engine with registry instance for update password in htpasswd
type testUsersRegistryAdapter struct {
	ctx     context.Context
	filters engine.QueryFilter
}

func newTestUsersRegistryAdapter(ctx context.Context, filters engine.QueryFilter) *testUsersRegistryAdapter {
	return &testUsersRegistryAdapter{
		ctx:     ctx,
		filters: filters,
	}
}

func (ra *testUsersRegistryAdapter) Users(getHtUsersFn UsersFn) ([]store.User, error) {
	if getHtUsersFn == nil {
		return nil, errors.New("userFn func undefined")
	}
	result, err := getHtUsersFn()
	if err != nil {
		return nil, err
	}

	var users = make([]store.User, 0)
	for user, passwd := range result {
		u := store.User{
			Name:     user,
			Password: string(passwd),
		}
		users = append(users, u)
	}

	if len(users) > 0 {
		return users, nil
	}

	return nil, errors.New("users list is empty")
}
