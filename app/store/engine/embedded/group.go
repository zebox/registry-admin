package embedded

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
)

var (

	// ErrRequiredFieldInGroupIsEmpty report about group name shouldn't be empty
	ErrRequiredFieldInGroupIsEmpty = errors.New("empty group name not allowed")

	// ErrFailedToCreateGroup report about error when group creating
	ErrFailedToCreateGroup = errors.New("failed to create new group")
)

// CreateGroup create a new group record
func (e *Embedded) CreateGroup(ctx context.Context, group *store.Group) (err error) {

	// check required parameters filled
	if group.Name == "" {
		return ErrRequiredFieldInGroupIsEmpty
	}

	createGroupSQL := fmt.Sprintf(`INSERT INTO %s (
		name,
		description
	) values(?, ?)`, groupsTable)
	stmt, err := e.db.PrepareContext(ctx, createGroupSQL)
	if err != nil {
		return ErrFailedToCreateGroup
	}
	defer func() { _ = stmt.Close() }()
	result, err := stmt.ExecContext(ctx, group.Name, group.Description)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		group.ID = id
	}
	return err
}

// GetGroup get data by group ID
func (e *Embedded) GetGroup(ctx context.Context, groupID int64) (group store.Group, err error) {

	queryFilter := fmt.Sprintf("select id,name,description from %s where id = ?", groupsTable)
	stmt, err := e.db.PrepareContext(ctx, queryFilter)
	if err != nil {
		return group, errors.Wrap(err, "failed to prepare query for group get")
	}
	defer func() { _ = stmt.Close() }()

	rows, err := stmt.QueryContext(ctx, groupID)
	if err != nil {
		return group, errors.Wrap(err, "failed to prepare query for group get")
	}
	defer func() { _ = rows.Close() }()

	emptyResult := true
	for rows.Next() {
		if err = rows.Scan(&group.ID, &group.Name, &group.Description); err != nil {
			return group, errors.Wrap(err, "failed scan group data")
		}
		emptyResult = false
	}
	if emptyResult {
		return group, errors.New("group record not found")
	}
	return group, nil
}

// FindGroups fetch list of existed group
func (e *Embedded) FindGroups(ctx context.Context, filter engine.QueryFilter) (groups engine.ListResponse, err error) {
	f := filtersBuilder(filter, "name")                                                            // set key filed for search query
	queryString := fmt.Sprintf("SELECT id,name,description FROM %s %s", groupsTable, f.allClauses) //nolint:gosec // query sanitizing calling before

	// avoid error shadowed
	var (
		stmt *sql.Stmt
		rows *sql.Rows
	)
	stmt, err = e.db.PrepareContext(ctx, queryString)
	if err != nil {
		return groups, err
	}

	rows, err = stmt.QueryContext(ctx)
	if err != nil {
		return groups, errors.Wrap(err, "failed to get groups list")
	}
	defer func() {
		_ = rows.Close()
	}()

	if groups.Total = e.getTotalRecordsExcludeRange(groupsTable, filter, []string{"name"}); groups.Total == 0 {
		return groups, nil // may be error handler catch
	}
	groups.Data = []interface{}{}
	for rows.Next() {
		var group store.Group
		if err = rows.Scan(&group.ID, &group.Name, &group.Description); err != nil {
			return groups, errors.Wrap(err, "failed scan group data")
		}
		groups.Data = append(groups.Data, group)
	}

	return groups, nil
}

// UpdateGroup update group records data
func (e *Embedded) UpdateGroup(ctx context.Context, group store.Group) (err error) {
	queryString := fmt.Sprintf(`UPDATE %s SET name='%s', description='%s' WHERE id = %d`, groupsTable, group.Name, group.Description, group.ID) //nolint:gosec // query sanitizing calling before
	res, err := e.db.ExecContext(ctx, queryString)
	if err != nil {
		return errors.Wrap(err, "failed to update user data")
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

// DeleteGroup delete users group record by ID
func (e *Embedded) DeleteGroup(ctx context.Context, id int64) (err error) {
	res, err := e.db.ExecContext(ctx, "DELETE FROM groups WHERE id = ?", id)
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
