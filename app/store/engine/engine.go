package engine

// Package engine defines interfaces each supported storage should implement.

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/zebox/registry-admin/app/store"
	"net/url"
	"regexp"
	"strconv"
)

// RepositoriesByUserAccess allow filtered repositories result list by assigned user access
// it's relevant only for role 'user' only
const (
	RepositoriesByUserAccess = "access.owner_id"
	AnonymousUserID          = int64(-1000) // uses for share access for anonymous and not registered users
	RegisteredUserID         = int64(-999)  // uses for share access for registered users only
)

type engineOptionsCtx string

// ErrNotFound return empty result error with request
var ErrNotFound = errors.New("record not found")

// ListResponse is a container for return list of result data
type ListResponse struct {
	Total int64         `json:"total"`
	Data  []interface{} `json:"data"`
}

// Interface defines methods provided by low-level storage engine
type Interface interface {

	// CreateUser create a new user login record in application storage
	CreateUser(ctx context.Context, user *store.User) (err error)

	// GetUser get data by user ID, where id can be any type
	GetUser(ctx context.Context, id interface{}) (user store.User, err error)

	// FindUsers fetch list of user by filter values.
	// @withPassword defines include or not hash value of password to result
	FindUsers(ctx context.Context, filter QueryFilter, withPassword bool) (users ListResponse, err error)

	// UpdateUser update user records data in application storage
	UpdateUser(ctx context.Context, user store.User) (err error)

	// DeleteUser delete user record by ID
	DeleteUser(ctx context.Context, id int64) (err error)

	// CreateGroup create a new users group record in application storage
	CreateGroup(ctx context.Context, group *store.Group) (err error)

	// GetGroup get data of users group by ID
	GetGroup(ctx context.Context, id int64) (group store.Group, err error)

	// FindGroups fetch list of existed users group
	FindGroups(ctx context.Context, filter QueryFilter) (groups ListResponse, err error)

	// UpdateGroup update data of users group
	UpdateGroup(ctx context.Context, group store.Group) (err error)

	// DeleteGroup delete users group record by ID
	DeleteGroup(ctx context.Context, id int64) (err error)

	// CreateAccess create a new access record for a specific user and repository
	CreateAccess(ctx context.Context, access *store.Access) (err error)

	// GetAccess get data of access entry by ID
	GetAccess(ctx context.Context, id int64) (access store.Access, err error)

	// FindAccesses get list of users access
	FindAccesses(ctx context.Context, filter QueryFilter) (accesses ListResponse, err error)

	// UpdateAccess will update a user access record
	UpdateAccess(ctx context.Context, access store.Access) (err error)

	// DeleteAccess delete access record by ID
	DeleteAccess(ctx context.Context, key string, id interface{}) (err error)

	// AccessGarbageCollector check outdated repositories in repositories table and delete ones from access list
	AccessGarbageCollector(ctx context.Context) error

	// CreateRepository create a new repository record
	CreateRepository(ctx context.Context, entry *store.RegistryEntry) (err error)

	// GetRepository get repository data by ID
	GetRepository(ctx context.Context, entryID int64) (entry store.RegistryEntry, err error)

	// FindRepositories get list of repositories
	FindRepositories(ctx context.Context, filter QueryFilter) (entries ListResponse, err error)

	// UpdateRepository update repository entry data
	UpdateRepository(ctx context.Context, conditionClause, data map[string]interface{}) (err error)

	// DeleteRepository delete repository entry by ID
	DeleteRepository(ctx context.Context, key string, id interface{}) (err error)

	// RepositoryGarbageCollector deletes outdated repositories entries
	RepositoryGarbageCollector(ctx context.Context, syncDate int64) (err error)

	// Close connection to storage instance
	Close(ctx context.Context) error
}

// QueryFilter using for query to data from storage
type QueryFilter struct {
	Range [2]int64 // array indexes are: 0 - Skip value, 1 - Limit value
	IDs   []int64  `json:"id"`

	// 'q' - key in filter use for full text search by fields which defined with parameters in filtersBuilder
	// other filters keys/values applies as exactly condition in query (at where clause)
	Filters map[string]interface{}

	Sort []string // ASC or DESC

	// GroupByFiled set field name to make a unique search by repositories table which contain duplicate repository names linked to different tags
	// GroupByField value for grouping should define in an engine implementation
	GroupByField bool
}

// FilterFromURLExtractor extracts param from URL and pass it to query which manipulation data in storage
func FilterFromURLExtractor(requestedURL *url.URL) (filters QueryFilter, err error) {
	_range, isRange := requestedURL.Query()["range"]
	sort, isSort := requestedURL.Query()["sort"]
	search, isSearch := requestedURL.Query()["filter"]

	// check and try to extract IDs from search string
	if isSearch {
		var query map[string]interface{}

		// check and try to extract strong condition by fields name
		err = json.Unmarshal([]byte(search[0]), &query)
		if len(query) > 0 && err == nil {
			filters.Filters = query
		}
	}

	// extract and parse range and sort params
	if isRange || isSort {
		rng, errRange := getRange(_range[0])
		if errRange != nil {
			return filters, errRange
		}
		filters.Range = rng
		filters.Sort = getQuotedStrings(sort[0])[:2]
	}

	return filters, err
}

// getRange parse URL search string param for store query filter
func getQuotedStrings(s string) []string {
	var re = regexp.MustCompile(`".*?"`)
	ms := re.FindAllString(s, -1)
	ss := make([]string, len(ms))
	for i, m := range ms {
		ss[i] = m[1 : len(m)-1]
	}
	return ss

}

// getRange parse URL range param for store query filter
func getRange(sRange string) (r [2]int64, err error) {
	var re = regexp.MustCompile(`(?m)\[(.*?),(.*?)]`)
	match := re.FindStringSubmatch(sRange)

	if len(match) == 3 {
		first, err := strconv.Atoi(match[1])
		if err != nil {
			return r, err
		}
		last, err := strconv.Atoi(match[2])
		if err != nil {
			return r, err
		}
		r[0], r[1] = int64(first), int64(last)+1 // +1 because js want range with start ZERO(0) index, but skip/limit DB function start from ONE(1)
	}
	return r, nil
}

const adminDefaultPasswordKey = "admin_default_key"

// SetAdminDefaultPassword allows defining default password for user admin when database with users created first
func SetAdminDefaultPassword(ctx context.Context, passwd *string) context.Context {
	newCtx := context.WithValue(ctx, engineOptionsCtx(adminDefaultPasswordKey), *passwd)
	*passwd = "" // clear default password from runtime memory
	return newCtx
}

// GetAdminDefaultPassword allows get default password for user admin from context
func GetAdminDefaultPassword(ctx context.Context) string {
	p := ctx.Value(engineOptionsCtx(adminDefaultPasswordKey))
	if p != nil {
		return p.(string)
	}
	return ""
}
