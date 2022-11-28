package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zebox/registry-admin/app/registry"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	R "github.com/go-pkgz/rest"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
)

// userHandlers implement controllers which allow manipulation with users model using REST API endpoints
type userHandlers struct {
	endpointsHandler
	registryService registryInterface
}

// usersRegistryAdapter need for bind FindUsers func in store engine with registry instance
// for update password when htpasswd is used
type usersRegistryAdapter struct {
	ctx     context.Context
	filters engine.QueryFilter
	usersFn registry.UsersFn
}

func newUsersRegistryAdapter(ctx context.Context, filters engine.QueryFilter, usersFunc registry.UsersFn) *usersRegistryAdapter {
	return &usersRegistryAdapter{
		ctx:     ctx,
		filters: filters,
		usersFn: usersFunc,
	}
}

func (ura *usersRegistryAdapter) Users() ([]store.User, error) {
	result, err := ura.usersFn(ura.ctx, ura.filters)
	if err != nil {
		return nil, err
	}

	var users []store.User
	for _, u := range result.Data {
		users = append(users, u.(store.User))
	}

	if len(users) > 0 {
		return users, nil
	}

	return nil, errors.New("users list is empty")
}

func (u *userHandlers) userCreateCtrl(w http.ResponseWriter, r *http.Request) {

	user := store.User{}
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "failed to parse user data for create with api")
		return
	}
	defer func() { _ = r.Body.Close() }()

	if user.Login == "" || user.Password == "" {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "empty login or password not allowed")
		return
	}

	err = u.dataStore.CreateUser(r.Context(), &user)
	if err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "failed create user with api")
		return
	}

	R.RenderJSON(w, responseMessage{Error: false, Message: "user created", ID: user.ID, Data: user})

	if err = u.registryService.UpdateHtpasswd(newUsersRegistryAdapter(r.Context(), engine.QueryFilter{}, u.dataStore.FindUsers)); err != nil {
		u.l.Logf("failed to update htpasswd: %v", err)
	}

}

func (u *userHandlers) userInfoCtrl(w http.ResponseWriter, r *http.Request) {

	userID := chi.URLParam(r, "id")

	// userInfo handler allows fetch user data only by user id
	i, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, u.l, http.StatusBadRequest, err, "failed to parse user id with api")
		return
	}

	user, err := u.dataStore.GetUser(r.Context(), i)
	if err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "failed get user with api")
		return
	}

	// password and it hashes shouldn't return with api
	user.Password = ""
	R.RenderJSON(w, responseMessage{
		Error: false,
		ID:    user.ID,
		Data:  user,
	})
}

func (u *userHandlers) userFindCtrl(w http.ResponseWriter, r *http.Request) {
	filter, err := engine.FilterFromUrlExtractor(r.URL)
	if err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "failed to parse URL parameters for make query filter")
		return
	}
	result, err := u.dataStore.FindUsers(r.Context(), filter)
	if err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "failed to find users")
		return
	}
	w.Header().Add("Content-Range", fmt.Sprintf("users %d-%d/%d", filter.Range[0], filter.Range[1], result.Total))
	R.RenderJSON(w, result)
}

func (u *userHandlers) userUpdateCtrl(w http.ResponseWriter, r *http.Request) { //nolint dupl
	userID := chi.URLParam(r, "id")

	i, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, u.l, http.StatusBadRequest, err, "failed to parse user id with api")
		return
	}

	user, err := u.dataStore.GetUser(r.Context(), i)
	if err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "failed to get user with api")
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&user); err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "failed to decode user data for update with api")
		return
	}

	if err = u.dataStore.UpdateUser(r.Context(), user); err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "failed to update user data with api")
		return
	}

	R.RenderJSON(w, responseMessage{
		Error:   false,
		Message: "ok",
		ID:      user.ID,
		Data:    user,
	})

	if err = u.registryService.UpdateHtpasswd(newUsersRegistryAdapter(r.Context(), engine.QueryFilter{}, u.dataStore.FindUsers)); err != nil {
		u.l.Logf("failed to update htpasswd: %v", err)
	}
}

func (u *userHandlers) userDeleteCtrl(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, u.l, http.StatusBadRequest, err, "failed to parse user id with api")
		return
	}

	if err = u.dataStore.DeleteUser(r.Context(), id); err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "failed to delete user with api")
		return
	}

	if err = u.dataStore.DeleteAccess(r.Context(), "owner_id", id); err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, fmt.Sprintf("failed to delete accesses for deleted user with id - '%d'", id))
		return
	}

	R.RenderJSON(w, responseMessage{Message: "user deleted"})

	if err = u.registryService.UpdateHtpasswd(newUsersRegistryAdapter(r.Context(), engine.QueryFilter{}, u.dataStore.FindUsers)); err != nil {
		u.l.Logf("failed to update htpasswd: %v", err)
	}
}
