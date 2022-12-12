package embedded

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
)

// CreateUser create a new user record
func (e *Embedded) CreateUser(ctx context.Context, user *store.User) (err error) {

	var emptyParams []string

	// check required parameters filled
	if user.Login == "" {
		emptyParams = append(emptyParams, "Login")
	}

	if user.Name == "" {
		emptyParams = append(emptyParams, "Name")
	}
	if user.Password == "" {
		emptyParams = append(emptyParams, "Password")
	}

	if !store.CheckRoleInList(user.Role) {
		emptyParams = append(emptyParams, fmt.Sprintf("role '%s' not allowed", user.Role))
	}

	if len(emptyParams) > 0 {
		return fmt.Errorf("required user fields not set: %s", strings.Join(emptyParams, ", "))
	}

	if user.Group == 0 {
		user.Group = 1
	}

	// hashing password value
	if errHash := user.HashAndSalt(); errHash != nil {
		return errHash
	}

	createUserSQL := fmt.Sprintf(`INSERT INTO %s (
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
		return multierror.Append(err, errors.New("failed to add new user"))
	}
	defer func() { _ = stmt.Close() }()
	result, err := stmt.ExecContext(ctx, user.Login, user.Name, user.Password, user.Role, user.Group, user.Disabled, user.Description)
	if err != nil {
		return multierror.Append(err, errors.New("failed to add new user"))
	}

	id, err := result.LastInsertId()
	if err == nil {
		user.ID = id
	}
	return err
}

// GetUser get data by user ID
func (e *Embedded) GetUser(ctx context.Context, id interface{}) (user store.User, err error) {
	var queryString string

	switch val := id.(type) {
	case string:
		// cast ID value when ID has login value
		queryString = fmt.Sprintf("select id,login,name,password,role,user_group,disabled,description from %s where login = ?", usersTable)

		// cast ID value when ID as string type
		if _, errParse := strconv.ParseInt(val, 10, 64); errParse == nil {
			queryString = fmt.Sprintf("select id,login,name,password,role,user_group,disabled,description from %s where id = ?", usersTable)
		}
	case int, int64:
		queryString = fmt.Sprintf("select id,login,name,password,role,user_group,disabled,description from %s where id = ?", usersTable)
	default:
		return user, errors.New("unsupported id type")
	}

	stmt, errPrep := e.db.PrepareContext(ctx, queryString)
	if errPrep != nil {
		return user, multierror.Append(err, errors.New("failed to get user"))
	}

	rows, err := stmt.QueryContext(ctx, id)
	if err != nil {
		return user, multierror.Append(err, errors.New("failed to get user"))
	}
	defer func() {
		_ = rows.Close()
	}()

	emptyResult := true
	for rows.Next() {
		if err = rows.Scan(&user.ID, &user.Login, &user.Name, &user.Password, &user.Role, &user.Group, &user.Disabled, &user.Description); err != nil {
			return user, multierror.Append(err, errors.New("failed scan user data"))
		}
		emptyResult = false
	}
	if emptyResult {
		return user, errors.New("record not found")
	}
	return user, nil
}

// FindUsers fetch list of user by filter values
func (e *Embedded) FindUsers(ctx context.Context, filter engine.QueryFilter) (users engine.ListResponse, err error) {
	f := filtersBuilder(filter, "login", "name")
	queryString := fmt.Sprintf("SELECT id,login,name,password,role,user_group,disabled,description FROM %s %s", usersTable, f.allClauses) //nolint:gosec // query sanitizing calling before

	// avoid error shadowed
	var (
		stmt *sql.Stmt
		rows *sql.Rows
	)
	stmt, err = e.db.PrepareContext(ctx, queryString)
	if err != nil {
		return users, err
	}

	rows, err = stmt.QueryContext(ctx)
	if err != nil {
		return users, errors.Wrap(err, "failed to get users list")
	}
	defer func() {
		_ = rows.Close()
	}()

	if users.Total = e.getTotalRecordsExcludeRange(usersTable, filter, []string{"login", "password"}); users.Total == 0 {
		return users, nil // may be error handler catch
	}
	users.Data = []interface{}{}
	for rows.Next() {
		var user store.User
		if err = rows.Scan(&user.ID, &user.Login, &user.Name, &user.Password, &user.Role, &user.Group, &user.Disabled, &user.Description); err != nil {
			return users, errors.Wrap(err, "failed scan user data")
		}
		user.Password = "" // clear password value when user fetch
		users.Data = append(users.Data, user)
	}

	return users, nil
}

// UpdateUser update user records data
func (e *Embedded) UpdateUser(ctx context.Context, user store.User) (err error) {

	if !store.CheckRoleInList(user.Role) {
		return errors.Errorf("role '%s' not allowed", user.Role)
	}
	var res sql.Result
	if user.Password != "" {
		if errHash := user.HashAndSalt(); errHash != nil {
			return errHash
		}
		res, err = e.db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET name=?, password=?, role=?, user_group=?, disabled=?, description=? WHERE id = ?", usersTable),
			user.Name, user.Password, user.Role, user.Group, user.Disabled, user.Description, user.ID)

	} else {
		// skip a password field update if updating password value is empty
		res, err = e.db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET name=?, role=?, user_group=?, disabled=?, description=? WHERE id = ?", usersTable),
			user.Name, user.Role, user.Group, user.Disabled, user.Description, user.ID)
	}

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

// DeleteUser delete user record by ID
func (e *Embedded) DeleteUser(ctx context.Context, id int64) (err error) {
	res, err := e.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
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
