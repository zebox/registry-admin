package embedded

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"strings"
)

// CreateRepository create a new repository record
func (e *Embedded) CreateRepository(ctx context.Context, entry *store.RegistryEntry) (err error) {

	createGroupSQL := fmt.Sprintf(`INSERT INTO %s (
		repository_name,
		tag,
		digest,
		size,
		pull_counter,
		timestamp,
		raw
	) values(?, ?, ?, ?, ?, ?, ?)`, repositoriesTable)
	stmt, err := e.db.PrepareContext(ctx, createGroupSQL)
	if err != nil {
		return errors.Wrap(err, "failed to create repository entry")
	}
	defer func() { _ = stmt.Close() }()
	result, err := stmt.ExecContext(ctx, entry.RepositoryName, entry.Tag, entry.Digest, entry.Size, entry.PullCounter, entry.Timestamp, entry.Raw)
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
func (e *Embedded) GetRepository(ctx context.Context, entryID int64) (entry store.RegistryEntry, err error) {

	queryFilter := fmt.Sprintf("SELECT id, repository_name, tag, digest, size, pull_counter, timestamp,raw FROM %s WHERE id = ?", repositoriesTable)
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
		if err = rows.Scan(&entry.ID, &entry.RepositoryName, &entry.Tag, &entry.Digest, &entry.Size, &entry.PullCounter, &entry.Timestamp, &entry.Raw); err != nil {
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

	queryString := fmt.Sprintf("SELECT * FROM %s %s", repositoriesTable, f.where) //nolint:gosec // query sanitizing calling before

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
		if err = rows.Scan(&entry.ID, &entry.RepositoryName, &entry.Tag, &entry.Digest, &entry.Size, &entry.PullCounter, &entry.Timestamp, &entry.Raw); err != nil {
			return entries, errors.Wrap(err, "failed scan repository data")
		}
		entries.Data = append(entries.Data, entry)
	}

	return entries, nil
}

// UpdateRepository update repository entry data
func (e *Embedded) UpdateRepository(ctx context.Context, conditionClause, data map[string]interface{}) (err error) {
	// filled fields set for sql query
	var fields []string
	for k, v := range data {
		fields = append(fields, fmt.Sprintf("%s=%s", k, castValueTypeToString(v)))
	}
	fieldSet := strings.Join(fields, ", ")

	// parse WHERE clause keys and values
	var conditions []string
	for k, v := range conditionClause {
		conditions = append(conditions, fmt.Sprintf("%s=%s", k, castValueTypeToString(v)))
	}
	conditionSet := strings.Join(conditions, " AND ")
	queryString := fmt.Sprintf(`UPDATE %s SET %s WHERE %s`, repositoriesTable, fieldSet, conditionSet) //nolint:gosec

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
func (e *Embedded) DeleteRepository(ctx context.Context, key string, id interface{}) (err error) {
	deleteSql := fmt.Sprintf("DELETE FROM %s WHERE %s=?", repositoriesTable, key) //nolint:gosec
	res, err := e.db.ExecContext(ctx, deleteSql, id)
	if err != nil {
		return errors.Wrapf(err, "failed execute query for user delete")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return err
}

// DeleteRepositoryGarbage deletes outdated repositories
func (e *Embedded) DeleteRepositoryGarbage(ctx context.Context, syncDate int64) (err error) {

	deleteSql := fmt.Sprintf("DELETE FROM %s WHERE %s!=?", repositoriesTable, store.RegistryTimestampField) //nolint:gosec
	res, err := e.db.ExecContext(ctx, deleteSql, syncDate)
	if err != nil {
		return errors.Wrapf(err, "failed execute query for user delete")
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return err
}
