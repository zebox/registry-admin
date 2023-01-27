package embedded

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"log"
	"strings"
)

// CreateRepository create a new repository record
func (e *Embedded) CreateRepository(ctx context.Context, entry *store.RegistryEntry) (err error) {

	createRepositorySQL := fmt.Sprintf(`INSERT INTO %s (
		repository_name,
		tag,
		digest,
		config_digest,
		size,
		pull_counter,
		timestamp,
		raw
	) values(?, ?, ?, ?, ?, ?, ?, ?)`, repositoriesTable)
	stmt, err := e.db.PrepareContext(ctx, createRepositorySQL)
	if err != nil {
		return errors.Wrap(err, "failed to create repository entry")
	}
	defer func() { _ = stmt.Close() }()
	result, err := stmt.ExecContext(ctx, entry.RepositoryName, entry.Tag, entry.Digest, entry.ConfigDigest, entry.Size, entry.PullCounter, entry.Timestamp, entry.Raw)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		entry.ID = id
	}
	return err
}

// GetRepository get repository data by ID
func (e *Embedded) GetRepository(ctx context.Context, entryID int64) (entry store.RegistryEntry, err error) { //nolint dupl

	queryFilter := fmt.Sprintf("SELECT id, repository_name, tag, digest, config_digest, size, pull_counter, timestamp,raw FROM %s WHERE id = ?", repositoriesTable)
	stmt, err := e.db.PrepareContext(ctx, queryFilter)
	if err != nil {
		return entry, errors.Wrap(err, "failed to prepare query for get repository data")
	}
	defer func() { _ = stmt.Close() }()

	rows, err := stmt.QueryContext(ctx, entryID)
	if err != nil {
		return entry, errors.Wrap(err, "failed to prepare query for get repository data")
	}
	defer func() { _ = rows.Close() }()

	emptyResult := true
	for rows.Next() {
		if err = rows.Scan(&entry.ID, &entry.RepositoryName, &entry.Tag, &entry.Digest, &entry.ConfigDigest, &entry.Size, &entry.PullCounter, &entry.Timestamp, &entry.Raw); err != nil {
			return entry, errors.Wrap(err, "failed scan group data")
		}
		emptyResult = false
	}
	if emptyResult {
		return entry, errors.New("repository entry not found")
	}
	return entry, nil
}

// FindRepositories fetch list of existed repositories
func (e *Embedded) FindRepositories(ctx context.Context, filter engine.QueryFilter) (entries engine.ListResponse, err error) {

	f := filtersBuilder(filter, "repository_name", "tag") // set key filed for search query

	// It needs for check request for 'groupBy', that show repositories list.
	// When request has 'groupBy' you should calculate summary size for each repository entry.
	// For request which show repositories entry (tags list) summary size not required
	sizeAggregateCheckerFn := func(isGroupBy bool) string {
		if isGroupBy {
			return "SUM(size)"
		}
		return "size"
	}

	//nolint:gosec // query sanitizing calling before
	queryString := fmt.Sprintf(
		"SELECT id,repository_name,tag,digest,config_digest,"+
			sizeAggregateCheckerFn(filter.GroupByField)+
			",pull_counter,timestamp,raw FROM %s %s", repositoriesTable, f.allClauses,
	)

	// check for select repositories by user access
	if _, ok := filter.Filters["access.owner_id"]; ok {
		queryString = fmt.Sprintf("SELECT repositories.id as id,"+
			"repository_name,"+
			"tag,"+
			"digest,"+
			"config_digest,"+
			sizeAggregateCheckerFn(filter.GroupByField)+","+
			"pull_counter,"+
			"timestamp,"+
			"raw "+
			"FROM %s "+
			"INNER JOIN access on repositories.repository_name=access.resource_name %s",
			repositoriesTable, f.allClauses,
		)
	}

	// avoid error shadowed
	var (
		stmt *sql.Stmt
		rows *sql.Rows
	)
	stmt, err = e.db.PrepareContext(ctx, queryString)
	if err != nil {
		return entries, err
	}

	rows, err = stmt.QueryContext(ctx)
	if err != nil {
		return entries, errors.Wrap(err, "failed to get repositories list")
	}
	defer func() {
		_ = rows.Close()
	}()

	if entries.Total = e.getTotalRecordsExcludeRange(repositoriesTable, filter, []string{"repository_name", "tag"}); entries.Total == 0 {
		return entries, nil // may be error handler catch
	}
	entries.Data = []interface{}{}
	for rows.Next() {
		var entry store.RegistryEntry
		if err = rows.Scan(&entry.ID, &entry.RepositoryName, &entry.Tag, &entry.Digest, &entry.ConfigDigest, &entry.Size, &entry.PullCounter, &entry.Timestamp, &entry.Raw); err != nil {
			return entries, errors.Wrap(err, "failed scan repository data")
		}
		entries.Data = append(entries.Data, entry)
	}

	return entries, nil
}

// UpdateRepository update repository entry data
func (e *Embedded) UpdateRepository(ctx context.Context, conditionClause, data map[string]interface{}) (err error) {
	// filled fields set for sql query
	var fields = make([]string, 0)
	for k, v := range data {
		fields = append(fields, fmt.Sprintf("%s=%s", k, castValueTypeToString(v)))
	}
	fieldSet := strings.Join(fields, ", ")

	// parse WHERE clause keys and values
	var conditions = make([]string, 0)
	for k, v := range conditionClause {
		conditions = append(conditions, fmt.Sprintf("%s=%s", k, castValueTypeToString(v)))
	}
	conditionSet := strings.Join(conditions, " AND ")

	//nolint:gosec // WHERE clause is a combination of conditions which prepared from filter and can't be paste with ? value
	queryString := fmt.Sprintf(`UPDATE %s SET %s WHERE %s`, repositoriesTable, fieldSet, conditionSet)

	res, err := e.db.ExecContext(ctx, queryString)
	if err != nil {
		return errors.Wrap(err, "failed to update repository data")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("record didn't update")
	}

	return err
}

// DeleteRepository delete repository entry by ID
func (e *Embedded) DeleteRepository(ctx context.Context, repositoryName, digest string) (err error) {

	//nolint:gosec // key value not passed from user input and can be change in code only
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE %s = ? AND %s = ?", repositoriesTable, store.RegistryRepositoryNameField, store.RegistryContentDigestField)
	res, err := e.db.ExecContext(ctx, deleteSQL, repositoryName, digest)
	if err != nil {
		return errors.Wrapf(err, "failed execute query for user delete")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return engine.ErrNotFound
	}

	return err
}

// RepositoryGarbageCollector deletes outdated repositories
func (e *Embedded) RepositoryGarbageCollector(ctx context.Context, syncDate int64) (err error) {

	//nolint:gosec // SQL query prepares from static value inside code logic and doesn't pass key name from outside
	deleteSQL := fmt.Sprintf("DELETE FROM %s WHERE %s<?", repositoriesTable, store.RegistryTimestampField)
	res, err := e.db.ExecContext(ctx, deleteSQL, syncDate)
	if err != nil {
		return errors.Wrapf(err, "failed execute query for user delete")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows > 0 {
		log.Printf("repositories deleted: %d", rows)
	}
	return err
}
