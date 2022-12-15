package registry

import (
	"fmt"
	"github.com/zebox/registry-admin/app/store"
	"os"
	"path/filepath"
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
	Users() ([]store.User, error)
}

// update will call every time when access list will change
func (ht *htpasswd) update(users []store.User) error {
	ht.lock.Lock()
	defer ht.lock.Unlock()

	// check file for exist
	err := createHtpasswdFileIfNoExist(ht.path)
	if err != nil {
		return err
	}

	if errTruncate := os.Truncate(ht.path, 0); err != nil {
		return fmt.Errorf("failed to truncate file %s: %v", ht.path, errTruncate)
	}

	f, err := os.OpenFile(ht.path, os.O_WRONLY|os.O_CREATE, 0o0600)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
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
