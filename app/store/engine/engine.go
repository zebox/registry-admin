package engine

// Package engine defines interfaces each supported storage should implement.

import (
	"context"
	"encoding/json"
	log "github.com/go-pkgz/lgr"
	"github.com/zebox/registry-admin/app/store"
	"net/url"
	"regexp"
	"strconv"
)

// Filters using for query filtered data from store
type Filters struct {
	Range   [2]int64
	Filters map[string]interface{}
	Sort    []string
}

type ListResponse struct {
	Total int64         `json:"total"`
	Data  []interface{} `json:"data"`
}

// Interface defines methods provided by low-level storage engine
type Interface interface {
	// Users model manipulation
	CreateUser(ctx context.Context, user *store.User) (err error)
	GetUser(ctx context.Context, id interface{}) (user store.User, err error)
	FindUsers(ctx context.Context, filter QueryFilter) (users ListResponse, err error)
	UpdateUser(ctx context.Context, user store.User) (err error)
	DeleteUser(ctx context.Context, id int64) (err error)

	// Groups model manipulation

	CreateGroup(ctx context.Context, group *store.Group) (err error)
	GetGroup(ctx context.Context, id int64) (group store.Group, err error)
	FindGroups(ctx context.Context, filter QueryFilter) (groups ListResponse, err error)
	UpdateGroup(ctx context.Context, group store.Group) (err error)
	DeleteGroup(ctx context.Context, id int64) (err error)

	// Accesses model manipulation
	CreateAccess(ctx context.Context, access *store.Access) (err error)
	GetAccess(ctx context.Context, id int64) (access store.Access, err error)
	FindAccesses(ctx context.Context, filter QueryFilter) (accesses ListResponse, err error)
	UpdateAccess(ctx context.Context, access store.Access) (err error)
	DeleteAccess(ctx context.Context, id int64) (err error)

	// Misc storage function
	Close(ctx context.Context) error
}

// QueryFilter using for query to data from storage

type QueryFilter struct {
	Range [2]int64 // array indexes, 0 - Skip value, 1 - Limit value
	IDs   []int64  `json:"id"`

	// 'q' - key in filter use for full text search by fields which defined with parameters in filtersBuilder
	// other filters keys/values applies as exactly condition in query (at where clause)
	Filters map[string]interface{}

	Sort []string // ASC or DESC
}

// FilterFromUrlExtractor extracts param from URL and pass it to query which manipulation data in storage
func FilterFromUrlExtractor(url *url.URL) (filters QueryFilter, err error) {
	_range, isRange := url.Query()["range"]
	sort, isSort := url.Query()["sort"]
	search, isSearch := url.Query()["filter"]

	// check and try to extract IDs from search string
	if isSearch {
		var query map[string]interface{}
		if filters.IDs, err = checkIDsExist(search[0]); err != nil {
			log.Printf("[DEBUG] fetch ids list failed %v", err)
			return filters, err
		}

		// check and try to extract strong condition by fields name
		err = json.Unmarshal([]byte(search[0]), &query)
		if len(query) > 0 && err == nil {
			filters.Filters = query
		}
	}

	// extract and parse range and sort params
	if isRange && isSort {
		rng, err := getRange(_range[0])
		if err != nil {
			return filters, err
		}
		filters.Range = rng
		filters.Sort = getQuotedStrings(sort[0])[:2]
	}

	return filters, err
}

// checkIDsExist checking url params contain IDs for include in store query filter
func checkIDsExist(str string) (ids []int64, err error) {
	var Ids struct {
		Ids []int64 `json:"ids"`
	}

	if err = json.Unmarshal([]byte(str), &Ids); err != nil {
		return ids, err
	}
	return Ids.Ids, nil
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
		return r, err
	}
	return r, err
}
