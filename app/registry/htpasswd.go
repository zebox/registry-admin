package registry

import (
	"bufio"
	"fmt"
	"github.com/zebox/registry-admin/app/store"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// htpasswd instance allow dynamic update .htpasswd file which use where basic auth is selected

// htpasswd holds a path to a system .htpasswd file and the machinery to parse
// it. Only bcrypt hash entries are supported.
type htpasswd struct {
	// path to .htpasswd access file which define in settings
	path string

	lock sync.Mutex
}

// FetchUsers interface allows get users list from store engine in registry instance
type FetchUsers interface {
	Users(func() (map[string][]byte, error)) ([]store.User, error)
}

// update will call every time when access list will change
func (ht *htpasswd) update(users []store.User) error {
	ht.lock.Lock()
	// check file for exist
	err := createHtpasswdFileIfNoExist(ht.path)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(ht.path, os.O_WRONLY|os.O_CREATE, 0o0600)
	if err != nil {
		ht.lock.Unlock()
		return err
	}

	defer func() {
		_ = f.Close()
		ht.lock.Unlock()
	}()

	for _, user := range users {
		if _, err := f.WriteString(fmt.Sprintf("%s:%s\n", user.Login, user.Password)); err != nil {
			return err
		}
	}

	return nil
}

// createHtpasswdFile creates  .htpasswd file with an update user in case the file is missing
func createHtpasswdFileIfNoExist(path string) error {
	if f, err := os.Open(filepath.Clean(path)); err == nil {
		_ = f.Close()
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o0700); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Clean(path), os.O_WRONLY|os.O_CREATE, 0o0600)
	if err != nil {
		return fmt.Errorf("failed to open htpasswd path %s", err)
	}
	defer func() { _ = f.Close() }()

	_, err = fmt.Fprint(f, nil)
	return err
}

// parseHTPasswd parses the contents of htpasswd. This will read all the
// entries in the file. An error is returned if a syntax errors
// are encountered or if the reader fails.
func (ht *htpasswd) parseHTPasswd() (map[string][]byte, error) {

	f, err := os.Open(ht.path)
	if err != nil {
		return nil, fmt.Errorf("failed to open .htpasswd file: %v", err)
	}

	entries := map[string][]byte{}
	scanner := bufio.NewScanner(f)
	var line int
	for scanner.Scan() {
		line++ // 1-based line numbering
		content := strings.TrimSpace(scanner.Text())

		if len(content) < 1 {
			continue
		}

		// lines that *begin* with a '#' are considered comments
		if content[0] == '#' {
			continue
		}

		i := strings.Index(content, ":")
		if i < 0 || i >= len(content) {
			return nil, fmt.Errorf("htpasswd: invalid entry at line %d: %q", line, scanner.Text())
		}

		entries[content[:i]] = []byte(content[i+1:])
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return entries, nil
}
