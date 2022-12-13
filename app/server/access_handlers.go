package server

import (
	"encoding/json"
	"fmt"
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

func (ah *accessHandlers) accessAddCtrl(w http.ResponseWriter, r *http.Request) {

	access := store.Access{}
	err := json.NewDecoder(r.Body).Decode(&access)
	if err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusInternalServerError, err, "failed to parse access data for create with api")
		return
	}
	defer func() { _ = r.Body.Close() }()

	err = ah.dataStore.CreateAccess(r.Context(), &access)
	if err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusInternalServerError, err, "failed add access with api")
		return
	}

	R.RenderJSON(w, responseMessage{Error: false, Message: "access added", ID: access.ID, Data: access})
}

func (ah *accessHandlers) accessInfoCtrl(w http.ResponseWriter, r *http.Request) {

	accessID := chi.URLParam(r, "id")

	// userInfo handler allows fetch user data only by user id
	i, err := strconv.ParseInt(accessID, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusBadRequest, err, "failed to parse access id with api")
		return
	}

	access, err := ah.dataStore.GetAccess(r.Context(), i)
	if err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusInternalServerError, err, "failed get access with api")
		return
	}

	R.RenderJSON(w, responseMessage{
		Error: false,
		ID:    access.ID,
		Data:  access,
	})
}

func (ah *accessHandlers) accessFindCtrl(w http.ResponseWriter, r *http.Request) {
	filter, err := engine.FilterFromURLExtractor(r.URL)
	if err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusInternalServerError, err, "failed to parse URL parameters for make query filter")
		return
	}
	result, err := ah.dataStore.FindAccesses(r.Context(), filter)
	if err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusInternalServerError, err, "failed to find access")
		return
	}
	w.Header().Add("Content-Range", fmt.Sprintf("accesses %d-%d/%d", filter.Range[0], filter.Range[1], result.Total))

	R.RenderJSON(w, result)
}

func (ah *accessHandlers) accessUpdateCtrl(w http.ResponseWriter, r *http.Request) { // nolint dupl
	groupID := chi.URLParam(r, "id")

	i, err := strconv.ParseInt(groupID, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusBadRequest, err, "failed to parse access id with api")
		return
	}

	access, err := ah.dataStore.GetAccess(r.Context(), i)
	if err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusInternalServerError, err, "failed to get access with api")
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&access); err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusInternalServerError, err, "failed to decode access data for update with api")
		return
	}

	if err = ah.dataStore.UpdateAccess(r.Context(), access); err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusInternalServerError, err, "failed to update access data with api")
		return
	}

	R.RenderJSON(w, responseMessage{
		Error: false,
		ID:    access.ID,
		Data:  access,
	})
}

func (ah *accessHandlers) accessDeleteCtrl(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusBadRequest, err, "failed to parse access id with api")
		return
	}

	if err := ah.dataStore.DeleteAccess(r.Context(), "id", id); err != nil {
		SendErrorJSON(w, r, ah.l, http.StatusInternalServerError, err, "failed to deleteDigest access with api")
		return
	}

	R.RenderJSON(w, responseMessage{Message: "access deleted"})
}
