package registry

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"
	"testing"
)

func TestRegistry_UpdateHtpasswd(t *testing.T) {
	var testUsers []store.User

	for i := 0; i < 10; i++ {
		user := store.User{
			Login:    fmt.Sprintf("user_%d", i),
			Password: fmt.Sprintf("password_%d", i),
		}
		require.NoError(t, user.HashAndSalt())
		testUsers = append(testUsers, user)
	}

	testPath := os.TempDir() + "/test/.htpasswd"

	r := Registry{htpasswd: &htpasswd{path: testPath}}
	require.NoError(t, r.UpdateHtpasswd(testUsers))

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

	assert.NoError(t, r.UpdateHtpasswd(testUsers))
	r.htpasswd.path = ""
	assert.Error(t, r.UpdateHtpasswd(testUsers))

	r.htpasswd = nil
	assert.Nil(t, r.UpdateHtpasswd(testUsers))
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
