package server

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	R "github.com/go-pkgz/rest"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"net/http"
	"strconv"
)

// accessHandlers implement controllers which allow manipulation with Access model using REST API endpoints
type accessHandlers struct {
	endpointsHandler
}

func (a *accessHandlers) accessAddCtrl(w http.ResponseWriter, r *http.Request) {

	access := store.Access{}
	err := json.NewDecoder(r.Body).Decode(&access)
	if err != nil {
		SendErrorJSON(w, r, a.l, http.StatusInternalServerError, err, "failed to parse access data for create with api")
		return
	}
	defer func() { _ = r.Body.Close() }()

	err = a.dataStore.CreateAccess(r.Context(), &access)
	if err != nil {
		SendErrorJSON(w, r, a.l, http.StatusInternalServerError, err, "failed add access with api")
		return
	}

	R.RenderJSON(w, responseMessage{Error: false, Message: "access added", ID: access.ID, Data: access})
}

func (a *accessHandlers) accessInfoCtrl(w http.ResponseWriter, r *http.Request) {

	accessId := chi.URLParam(r, "id")

	// userInfo handler allows fetch user data only by user id
	i, err := strconv.ParseInt(accessId, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, a.l, http.StatusBadRequest, err, "failed to parse access id with api")
		return
	}

	access, err := a.dataStore.GetAccess(r.Context(), i)
	if err != nil {
		SendErrorJSON(w, r, a.l, http.StatusInternalServerError, err, "failed get access with api")
		return
	}

	R.RenderJSON(w, responseMessage{
		Error: false,
		ID:    access.ID,
		Data:  access,
	})
}

func (a *accessHandlers) accessFindCtrl(w http.ResponseWriter, r *http.Request) {
	filter, err := engine.FilterFromUrlExtractor(r.URL)
	if err != nil {
		SendErrorJSON(w, r, a.l, http.StatusInternalServerError, err, "failed to parse URL parameters for make query filter")
		return
	}
	result, err := a.dataStore.FindAccesses(r.Context(), filter)
	if err != nil {
		SendErrorJSON(w, r, a.l, http.StatusInternalServerError, err, "failed to find access")
		return
	}

	R.RenderJSON(w, result)
}

func (a *accessHandlers) accessUpdateCtrl(w http.ResponseWriter, r *http.Request) { // nolint dupl
	groupId := chi.URLParam(r, "id")

	i, err := strconv.ParseInt(groupId, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, a.l, http.StatusBadRequest, err, "failed to parse access id with api")
		return
	}

	access, err := a.dataStore.GetAccess(r.Context(), i)
	if err != nil {
		SendErrorJSON(w, r, a.l, http.StatusInternalServerError, err, "failed to get access with api")
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&access); err != nil {
		SendErrorJSON(w, r, a.l, http.StatusInternalServerError, err, "failed to decode access data for update with api")
		return
	}

	if err = a.dataStore.UpdateAccess(r.Context(), access); err != nil {
		SendErrorJSON(w, r, a.l, http.StatusInternalServerError, err, "failed to update access data with api")
		return
	}

	R.RenderJSON(w, responseMessage{
		Error: false,
		ID:    access.ID,
		Data:  access,
	})
}

func (a *accessHandlers) accessDeleteCtrl(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, a.l, http.StatusBadRequest, err, "failed to parse access id with api")
		return
	}

	if err := a.dataStore.DeleteAccess(r.Context(), id); err != nil {
		SendErrorJSON(w, r, a.l, http.StatusInternalServerError, err, "failed to delete access with api")
		return
	}

	R.RenderJSON(w, responseMessage{Message: "access deleted"})
}
