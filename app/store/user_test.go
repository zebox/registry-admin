package store

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestUser_HashAndSalt(t *testing.T) {
	testUser := User{
		Login:    "test_login",
		Name:     "test_user",
		Password: "test_password",
	}

	err := testUser.HashAndSalt()
	require.NoError(t, err)
	hashCheckRegExp := regexp.MustCompile(`(?m)^\$2[ayb]\$.{56}$`)
	assert.True(t, hashCheckRegExp.MatchString(testUser.Password))

}

func TestComparePassword(t *testing.T) {

	testUser := User{
		Login:    "test_login",
		Name:     "test_user",
		Password: "test_password",
	}

	err := testUser.HashAndSalt()
	require.NoError(t, err)
	hashCheckRegExp := regexp.MustCompile(`(?m)^\$2[ayb]\$.{56}$`)
	assert.True(t, hashCheckRegExp.MatchString(testUser.Password))

	testUserForCheck := User{
		Login:    "test_login",
		Name:     "test_user",
		Password: "test_password",
	}

	assert.True(t, ComparePassword(testUser.Password, testUserForCheck.Password))

	testUserForCheck.Password = "fake_password"
	assert.False(t, ComparePassword(testUser.Password, testUserForCheck.Password))
}

func TestCheckRoleInList(t *testing.T) {
	for _, r := range roles {
		assert.True(t, CheckRoleInList(r))
	}
	assert.False(t, CheckRoleInList("fakeRoles"))
}
