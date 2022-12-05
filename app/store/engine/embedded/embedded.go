package embedded

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	_ "github.com/mattn/go-sqlite3"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
)

const (
	usersTable        = "users"
	groupsTable       = "groups"
	accessTable       = "access"
	repositoriesTable = "repositories"
)

var (
	ErrTableAlreadyExist = errors.New("table already exist or has an error")
)

type Embedded struct {
	Path string `json:"path"`
	db   *sql.DB
}

type queryFilter struct {
	skipLimit  string // an offset and a limit params
	order      string // an order by clause
	in         string // in array values
	where      string // where without limit and offset params, return all items when math where clause
	allClauses string // raw where clause with skip and limit
	groupBy    string
}

func NewEmbedded(pathToDB string) *Embedded {
	return &Embedded{Path: pathToDB}
}

func (e *Embedded) Connect(ctx context.Context) (err error) {

	e.db, err = sql.Open("sqlite3", e.Path)
	if err != nil || e.Path == "" {
		return err
	}

	// close connection global using context
	go func() {
		<-ctx.Done()
		_ = e.db.Close()
	}()
	return e.initTables(ctx)

}

func (e *Embedded) initTables(ctx context.Context) (err error) {
	if err = e.initUserTable(ctx); err != nil {
		err = multierror.Append(err, errors.Errorf("failed to create %s table", usersTable))
	}

	if err = e.initGroupsTable(ctx); err != nil {
		err = multierror.Append(err, errors.Errorf("failed to create %s table", groupsTable))
	}

	if err = e.initAccessTable(ctx); err != nil {
		err = multierror.Append(err, errors.Errorf("failed to create %s table", accessTable))
	}

	if err = e.initRepositoriesTable(ctx); err != nil {
		err = multierror.Append(err, errors.Errorf("failed to create %s table", repositoriesTable))
	}

	// SQLite driver doesn't catch error if file doesn't exist and try to create a new database file.
	// But if path which passed to drive has invalid path name SQLite doesn't throw error too.
	// Because check for file exist required after first write transaction (such create table or other)
	if _, errStat := os.Stat(e.Path); os.IsNotExist(errStat) {
		return fmt.Errorf("[ERROR] database path is invalid '%s'. Can't create database file", e.Path)
	}
	return err
}

func (e *Embedded) initGroupsTable(ctx context.Context) (err error) {
	if exist, err := e.isTableExist(ctx, groupsTable); err != nil || exist {
		return ErrTableAlreadyExist
	}

	sqlText := fmt.Sprintf(`CREATE TABLE %s(
	id    INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,	
	name TEXT UNIQUE,
	description TEXT)`, groupsTable)

	_, err = e.db.Exec(sqlText)
	if err != nil {
		return errors.Wrapf(err, "failed to create %s table", groupsTable)
	}

	// create default admin group when a new database creation
	group := store.User{
		Name:        "default", // default login
		Description: "Default administration group",
	}

	createUserSQL := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		name,
		description
	) values(?, ?)`, groupsTable)

	stmt, err := e.db.Prepare(createUserSQL)
	if err != nil {
		return errors.Wrapf(err, "failed to insert new default user to %s table", groupsTable)
	}
	defer func() { _ = stmt.Close() }()
	_, err = stmt.Exec(group.Name, group.Description)
	return err
}

func (e *Embedded) initUserTable(ctx context.Context) error {
	if exist, err := e.isTableExist(ctx, usersTable); err != nil || exist {
		return ErrTableAlreadyExist
	}

	sqlText := fmt.Sprintf(`CREATE TABLE %s(
	id    INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
	login  TEXT UNIQUE,
	name TEXT,
	password TEXT,
	role TEXT,
	user_group INTEGER,
	disabled INTEGER,
	description TEXT)`, usersTable)

	_, err := e.db.Exec(sqlText)
	if err != nil {
		return errors.Wrapf(err, "failed to create %s table", usersTable)
	}

	// create default admin user if new database creation
	user := store.User{
		Login:       "admin",
		Name:        "admin", // default login
		Password:    "admin", // default password
		Role:        "admin",
		Group:       1,
		Description: "Default user with administration role ",
	}

	// hashing password
	if err = user.HashAndSalt(); err != nil {
		return err
	}

	createUserSQL := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		login,
		name,
		password,
		role,
		user_group,
		disabled,
		description
		
	) values(?, ?, ?, ?, ?, ?, ?)`, usersTable)
	stmt, err := e.db.Prepare(createUserSQL)
	if err != nil {
		return errors.Wrapf(err, "failed to insert new default user to %s table", usersTable)
	}
	defer func() { _ = stmt.Close() }()
	_, err = stmt.Exec(user.Login, user.Name, user.Password, user.Role, user.Group, 0, user.Description)
	return err
}

func (e *Embedded) initAccessTable(ctx context.Context) (err error) {
	if exist, err := e.isTableExist(ctx, accessTable); err != nil || exist {
		return ErrTableAlreadyExist
	}

	sqlText := fmt.Sprintf(`CREATE TABLE %s(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		owner_id INTEGER NOT NULL,
		is_group INTEGER,
		name TEXT,
		resource_type TEXT,
		resource_name TEXT,
		action TEXT,
		disabled INTEGER,
		UNIQUE(owner_id,resource_type,resource_name,action))`, accessTable)

	_, err = e.db.Exec(sqlText)
	if err != nil {
		return multierror.Append(err, errors.Errorf("failed to create %s table", accessTable))
	}
	return nil
}

func (e *Embedded) initRepositoriesTable(ctx context.Context) (err error) {
	if exist, err := e.isTableExist(ctx, repositoriesTable); err != nil || exist {
		return ErrTableAlreadyExist
	}

	sqlText := fmt.Sprintf(`CREATE TABLE %s(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		repository_name INTEGER NOT NULL CHECK(repository_name <> ''),
		tag TEXT NOT NULL CHECK(tag <> ''),
		digest TEXT NOT NULL CHECK(digest <> ''),
		config_digest TEXT NOT NULL CHECK(config_digest <> ''),
		size INTEGER,
		pull_counter INTEGER,
		timestamp INTEGER,
		raw TEXT,
		UNIQUE(repository_name,tag))`, repositoriesTable)

	_, err = e.db.Exec(sqlText)
	if err != nil {
		return multierror.Append(err, errors.Errorf("failed to create %s table", repositoriesTable))
	}
	return nil
}

func (e *Embedded) isTableExist(_ context.Context, tableName string) (exist bool, err error) {

	rows, err := e.db.Query(fmt.Sprintf("select DISTINCT tbl_name from sqlite_master where tbl_name = '%s'", tableName))
	if err != nil {
		return false, multierror.Append(err, errors.Errorf("can't check for %s table exist", tableName))
	}

	defer func() { _ = rows.Close() }()
	for rows.Next() {
		return true, nil
	}
	return false, nil
}

func (e *Embedded) Close(_ context.Context) error {
	return e.db.Close()
}

// filtersBuilder parse an engine filter values and build query filter for 'embedded' implementation
// IMPORTANT: value for group by always fetch from FIRST index of 'fieldsName' list, keep in mind this when use 'group by'
func filtersBuilder(filter engine.QueryFilter, fieldsName ...string) (f queryFilter) {

	var ids string

	// skip and limit statement build
	skip := ""
	if filter.Range[0] > 0 {
		skip = fmt.Sprintf("OFFSET %d", filter.Range[0])
		f.skipLimit = skip
	}

	if filter.Range[1] > 0 {
		limit := fmt.Sprintf(" LIMIT %d", filter.Range[1]-filter.Range[0])
		f.skipLimit = fmt.Sprintf("%s %s", limit, skip)
	}

	var (
		like             string
		strongConditions []string
	)

	// search query statement and parse queryFilter value
	for k, v := range filter.Filters {

		// check sql value for sql-injection
		k, v = sanitizeKeyValue(k, v)

		// build filter by list of IDs
		if k == "ids" {
			var stringIds []string
			for _, value := range v.([]interface{}) {
				stringIds = append(stringIds, castValueTypeToString(value))
			}
			ids = strings.Join(stringIds, ", ")
			f.in = fmt.Sprintf("id IN (%s)", ids)
			continue
		}

		if k == "q" {
			var likeConndition []string
			for _, val := range fieldsName {
				if reflect.TypeOf(v).Kind() == reflect.Int {
					likeConndition = append(likeConndition, fmt.Sprintf(" %s LIKE %d", val, v))
					continue
				}
				likeConndition = append(likeConndition, fmt.Sprintf("%s LIKE '%%%s%%'", val, v))
			}
			like = strings.Join(likeConndition, " OR ")
			continue
		}

		conditionValue := fmt.Sprintf("%s = %s", k, castValueTypeToString(v))
		if k == engine.RepositoriesByUserAccess {
			conditionValue = fmt.Sprintf("access.owner_id = %s", castValueTypeToString(v))
		}

		strongConditions = append(strongConditions, conditionValue)
	}

	var strongCondition string
	if len(strongConditions) > 0 {
		if like != "" {
			strongCondition = fmt.Sprintf("AND (%s)", strings.Join(strongConditions, " AND "))
		} else {
			strongCondition = fmt.Sprintf("(%s)", strings.Join(strongConditions, " AND "))
		}

	}

	if f.in != "" {
		f.allClauses = fmt.Sprintf("WHERE %s", f.in)
	}

	if like != "" {
		if f.allClauses == "" {
			f.allClauses = fmt.Sprintf("WHERE (%s)", like)
		} else {
			f.allClauses = fmt.Sprintf("%s AND (%s)", f.allClauses, like)
		}

	}

	if strongCondition != "" {
		if f.allClauses == "" {
			f.allClauses = fmt.Sprintf("WHERE %s", strongCondition)
		} else {
			f.allClauses = fmt.Sprintf("%s %s", f.allClauses, strongCondition)
		}
	}
	f.where = f.allClauses

	if filter.Sort == nil {
		filter.Sort = []string{"id", "asc"} // default sorting
	}

	// set value for group by clause.
	// IMPORTANT: group by field name always fetch from first index of fieldsName option
	if filter.GroupByField && len(fieldsName) > 0 {
		f.groupBy = "GROUP BY " + fieldsName[0]
	}

	f.order = fmt.Sprintf("%s ORDER BY %s %s ", f.groupBy, filter.Sort[0], filter.Sort[1])

	f.allClauses = f.allClauses + f.order + f.skipLimit

	return f
}

// getTotalRecordExcludeRange return total number of records exclude range/skip clause for pagination support
// 		tableName - specify table name for search
//		filter - set of params for where clause in query
//		searchFields - define list of key fields using in where clause
func (e *Embedded) getTotalRecordsExcludeRange(tableName string, filter engine.QueryFilter, searchFields []string) int64 {
	filter.Range = [2]int64{0, 0} // clear skip/offset range

	// it defines request type for get total records from table with duplicates fields such like repositories
	countType := "COUNT(*)"
	if filter.GroupByField && len(searchFields) > 0 {
		filter.GroupByField = false // reset 'GROUP BY' clause for exclude it from filterBuilder

		// IMPORTANT: for distinct values gets FIRST item of searchFields list
		countType = fmt.Sprintf("COUNT(DISTINCT %s)", searchFields[0])
	}

	f := filtersBuilder(filter, searchFields...)
	queryString := fmt.Sprintf("SELECT %s FROM %s %s", countType, tableName, f.allClauses)

	// check for select repositories by user access
	if _, ok := filter.Filters["access.owner_id"]; ok {

		queryString = fmt.Sprintf("SELECT %s FROM %s INNER JOIN access on repositories.repository_name=access.resource_name %s", countType, tableName, f.where)
	}

	rows, err := e.db.Query(queryString)
	if err != nil {
		return 0
	}

	defer func() {
		_ = rows.Close()
	}()

	var recordsCounter int64
	rows.Next()
	if err = rows.Scan(&recordsCounter); err != nil {
		return 0
	}
	return recordsCounter
}

// castValueTypeToString will select appropriate type to formatting string
func castValueTypeToString(value interface{}) string {
	switch v := value.(type) {
	case string, digest.Digest, []uint8:
		return fmt.Sprintf("'%s'", v)
	case []string:
		if len(v) > 0 {
			return fmt.Sprintf("'%s'", v[0])
		}
	case int, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.f", v)
	}
	return ""
}

// sanitizeValue check key name and value for contain sql-injection code and cleanup ones
func sanitizeKeyValue(key string, value interface{}) (cleanKey string, cleanValue interface{}) {

	// query value input can be full text search string with white spaces and contain substring either 'OR' or 'AND'
	// because in this regexp white spaces, substrings contain 'OR' or/and 'AND' and not will be replaced
	// 'OR', 'AND' will replace if they wrap of white spaces
	var queryValueRegExp = regexp.MustCompile(`(?i)[\t\r\n]|(--)|(%)|\s{2,}|(\s(OR|AND|JOIN|LEFT|RIGHT|LIKE)\s)|\)|\(|'|"|=|\*|SELECT|UPDATE|INSERT|DELETE|LIKE|WHERE|ALTER|UNION`)

	// same regexp as above but include trim white spaces between words of string for ke or value
	var keyNameValueRegExp = regexp.MustCompile(`(?i)[\t\r\n]|(--)|\s+|(%)|(\sOR\s|\sAND\s|\)|\(|'|"|=|\*|SELECT|UPDATE|INSERT|DELETE|LIKE|WHERE|ALTER|UNION)`)

	// search sql-injection code in key name
	cleanKey = key
	for {
		isPatternDetected := false
		for _, match := range keyNameValueRegExp.FindAllString(cleanKey, -1) {
			cleanKey = strings.Replace(cleanKey, match, "", -1)
			isPatternDetected = true
		}
		if !isPatternDetected {
			break
		}
	}

	// search sql-injection code in value
	cleanValue = value
	switch val := value.(type) {
	case string:
		{

			tmpString := val

			for {
				isPatternDetected := false

				// full text query value string sanitizing
				if cleanKey == "q" {
					for _, match := range queryValueRegExp.FindAllString(tmpString, -1) {
						tmpString = strings.Replace(tmpString, match, "", -1)
						isPatternDetected = true
					}

					// sanitize a filter value
				} else {
					for _, match := range keyNameValueRegExp.FindAllString(tmpString, -1) {
						tmpString = strings.Replace(tmpString, match, "", -1)
						isPatternDetected = true
					}
				}

				if !isPatternDetected {
					cleanValue = tmpString
					break
				}
			}
		}
	}

	return cleanKey, cleanValue
}
