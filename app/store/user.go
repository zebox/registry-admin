package store

import (
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"log"
)

// List user roles
var (
	// roles define access right for different user roles
	// admin - has full access for entire registry end-points
	// manager - allow pull,push and delete for type 'repository'
	// a repository type defined with https://docs.docker.com/registry/spec/auth/scope/
	roles = []string{"admin", "manager", "user"}
)

// User holds user-related info
type User struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Login       string `json:"login"`
	Password    string `json:"password"`
	Role        string `json:"role"`  // role name selected by index from roles item
	Group       int64  `json:"group"` // reference to group ID
	Disabled    bool   `json:"blocked"`
	Description string `json:"description"`
}

// Group holds user group
type Group struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// HashAndSalt encrypted user password
func (u *User) HashAndSalt() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return multierror.Append(err, errors.New("failed to crypt user password"))
	}
	u.Password = string(hash)
	return nil
}

// ComparePassword checking password for match
func ComparePassword(passwordHash, passwordString string) bool {

	bPwdHash := []byte(passwordHash)
	bPwdString := []byte(passwordString)

	err := bcrypt.CompareHashAndPassword(bPwdHash, bPwdString)
	if err != nil {
		log.Printf("[ERROR] password doesn't match %v", err)
		return false
	}
	return true
}

// CheckRoleInList checking role assigned to user when add or update for exist in roles (allowed role list roles)
func CheckRoleInList(role string) bool {
	for _, existedRole := range roles {
		if role == existedRole {
			return true
		}
	}
	return false
}
