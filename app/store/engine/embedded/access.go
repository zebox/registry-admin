package embedded

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"strings"
)

// CreateAccess create a new user access record
func (e *Embedded) CreateAccess(ctx context.Context, access *store.Access) (err error) {

	var emptyParams []string

	// check required parameters filled
	if access.Name == "" {
		emptyParams = append(emptyParams, "Name")
	}

	if access.Type == "" {
		emptyParams = append(emptyParams, "Type")
	}

	if access.ResourceName == "" {
		emptyParams = append(emptyParams, "Resource name")
	}

	if access.Action == "" {
		emptyParams = append(emptyParams, "Action")
	}

	if access.Owner == 0 {
		emptyParams = append(emptyParams, "Owner")
	}

	if len(emptyParams) > 0 {
		return fmt.Errorf("required access fields not set: %s", strings.Join(emptyParams, ", "))
	}

	createAccessSQL := fmt.Sprintf(`INSERT INTO %s (
		owner_id,
		is_group,
		name,
		resource_type,
		resource_name,
		action,
		disabled
	) values (?, ?, ?, ?, ?, ?, ?)`, accessTable)
	stmt, err := e.db.PrepareContext(ctx, createAccessSQL)

	if err != nil {
		return errors.Wrap(err, "failed to add new access")
	}

	defer func() { _ = stmt.Close() }()
	result, err := stmt.ExecContext(ctx, access.Owner, access.IsGroup, access.Name, access.Type, access.ResourceName, access.Action, access.Disabled)
	if err != nil {
		return errors.Wrap(err, "failed to add new access")
	}

	id, err := result.LastInsertId()
	if err == nil {
		access.ID = id
	}
	return err
}

// GetAccess get data of access entry by ID
func (e *Embedded) GetAccess(ctx context.Context, id int64) (access store.Access, err error) {

	queryFilter := fmt.Sprintf("select * from %s where id = ?", accessTable)
	stmt, err := e.db.PrepareContext(ctx, queryFilter)
	if err != nil {
		return access, errors.Wrap(err, "failed to prepare query for group get")
	}
	defer func() { _ = stmt.Close() }()

	rows, err := stmt.QueryContext(ctx, id)
	if err != nil {
		return access, errors.Wrap(err, "failed to prepare query for group get")
	}
	defer func() { _ = rows.Close() }()

	emptyResult := true
	for rows.Next() {

		if err = rows.Scan(&access.ID, &access.Owner, &access.IsGroup, &access.Name, &access.Type, &access.ResourceName, &access.Action, &access.Disabled); err != nil {
			return access, errors.Wrap(err, "failed scan access data")
		}
		emptyResult = false
	}
	if emptyResult {
		return access, errors.New("access record not found")
	}
	return access, nil
}

// FindAccesses get list of existed users access
func (e *Embedded) FindAccesses(ctx context.Context, filter engine.QueryFilter) (accesses engine.ListResponse, err error) {
	f := filtersBuilder(filter, "name", "resource_name")
	queryString := fmt.Sprintf("SELECT * FROM %s %s", accessTable, f.allClauses) //nolint:gosec // query sanitizing calling before

	rows, err := e.db.QueryContext(ctx, queryString)
	if err != nil {
		return accesses, errors.Wrap(err, "failed to get access list")
	}
	defer func() {
		_ = rows.Close()
	}()
	accesses.Data = []interface{}{}

	if accesses.Total = e.getTotalRecordsExcludeRange(accessTable, filter, []string{"owner_id", "resource_name"}); accesses.Total == 0 {
		return accesses, nil // may be error handler catch
	}

	for rows.Next() {
		var tmpAccess store.Access
		if err = rows.Scan(&tmpAccess.ID, &tmpAccess.Owner, &tmpAccess.IsGroup, &tmpAccess.Name, &tmpAccess.Type, &tmpAccess.ResourceName, &tmpAccess.Action, &tmpAccess.Disabled); err != nil {
			return accesses, errors.Wrap(err, "failed scan access data")
		}
		accesses.Data = append(accesses.Data, tmpAccess)
	}

	return accesses, nil
}

// UpdateAccess will update access record
func (e *Embedded) UpdateAccess(ctx context.Context, access store.Access) (err error) {

	// fields order: owner_id, is_group, name, resource_type, resource_name, action, disabled
	res, err := e.db.ExecContext(ctx, "UPDATE access SET owner_id=?, is_group=?, name=?, resource_type=?, resource_name=?, action=?, disabled=? WHERE id = ?",
		access.Owner, access.IsGroup, access.Name, access.Type, access.ResourceName, access.Action, access.Disabled, access.ID)
	if err != nil {
		return errors.Wrap(err, "failed to update access data")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("record access didn't update")
	}

	return err
}

// DeleteAccess delete access record by ID
func (e *Embedded) DeleteAccess(ctx context.Context, key string, id interface{}) (err error) {

	//nolint:gosec // key value not passed from user input and can be change in code only
	query := fmt.Sprintf("DELETE FROM access WHERE %s = ?", key)
	res, err := e.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrapf(err, "failed execute query for access delete")
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

// AccessGarbageCollector check outdated repositories in repositories table and delete ones from access list
func (e *Embedded) AccessGarbageCollector(ctx context.Context) error {
	res, err := e.db.ExecContext(ctx, "DELETE FROM access WHERE resource_name NOT IN (SELECT repository_name FROM repositories)")
	if err != nil {
		return errors.Wrapf(err, "failed execute query for execute garbage collector for access ")
	}
	_, err = res.RowsAffected()
	if err != nil {
		return err
	}

	return err
}
