package server

import (
	"encoding/json"
	"fmt"
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

}

func (u *userHandlers) userInfoCtrl(w http.ResponseWriter, r *http.Request) {

	userId := chi.URLParam(r, "id")

	// userInfo handler allows fetch user data only by user id
	i, err := strconv.ParseInt(userId, 10, 64)
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
	userId := chi.URLParam(r, "id")

	i, err := strconv.ParseInt(userId, 10, 64)
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
		Error: false,
		ID:    user.ID,
		Data:  user,
	})
}

func (u *userHandlers) userDeleteCtrl(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, u.l, http.StatusBadRequest, err, "failed to parse user id with api")
		return
	}

	if err := u.dataStore.DeleteUser(r.Context(), id); err != nil {
		SendErrorJSON(w, r, u.l, http.StatusInternalServerError, err, "failed to delete user with api")
		return
	}

	R.RenderJSON(w, responseMessage{Message: "user deleted"})
}
