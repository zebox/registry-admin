package registry

import (
	"os"
	"sync"
)

// htpasswd instance allow dynamic update .htpasswd file which use where basic auth is selected

// htpasswd holds a path to a system .htpasswd file and the machinery to parse
// it. Only bcrypt hash entries are supported.
type htpasswd struct {
	// path to .htpasswd access file which define in setting
	path string

	lock sync.Mutex
}

// update will call every time when access list will change
func (ht htpasswd) update(map[string][]byte) error {

	// check file for exist
	_, err := os.Stat(ht.path)
	if err != nil {
		return err
	}

	return nil
}
