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

// groupHandlers implement controllers which allow manipulation with user group model using REST API endpoints
type groupHandlers struct {
	endpointsHandler
}

func (g *groupHandlers) groupCreateCtrl(w http.ResponseWriter, r *http.Request) {

	group := store.Group{}
	err := json.NewDecoder(r.Body).Decode(&group)
	if err != nil {
		SendErrorJSON(w, r, g.l, http.StatusInternalServerError, err, "failed to parse group data for create with api")
		return
	}
	defer func() { _ = r.Body.Close() }()

	err = g.dataStore.CreateGroup(r.Context(), &group)
	if err != nil {
		SendErrorJSON(w, r, g.l, http.StatusInternalServerError, err, "failed create group with api")
		return
	}

	R.RenderJSON(w, responseMessage{Error: false, Message: "group created", ID: group.ID, Data: group})

}

func (g *groupHandlers) groupInfoCtrl(w http.ResponseWriter, r *http.Request) {

	groupId := chi.URLParam(r, "id")

	// userInfo handler allows fetch user data only by user id
	i, err := strconv.ParseInt(groupId, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, g.l, http.StatusBadRequest, err, "failed to parse group id with api")
		return
	}

	group, err := g.dataStore.GetGroup(r.Context(), i)
	if err != nil {
		SendErrorJSON(w, r, g.l, http.StatusInternalServerError, err, "failed get group with api")
		return
	}

	R.RenderJSON(w, responseMessage{
		Error: false,
		ID:    group.ID,
		Data:  group,
	})
}

func (g *groupHandlers) groupFindCtrl(w http.ResponseWriter, r *http.Request) {
	filter, err := engine.FilterFromUrlExtractor(r.URL)
	if err != nil {
		SendErrorJSON(w, r, g.l, http.StatusInternalServerError, err, "failed to parse URL parameters for make query filter")
		return
	}
	result, err := g.dataStore.FindGroups(r.Context(), filter)
	if err != nil {
		SendErrorJSON(w, r, g.l, http.StatusInternalServerError, err, "failed to find groups")
		return
	}

	w.Header().Add("Content-Range", fmt.Sprintf("groups %d-%d/%d", filter.Range[0], filter.Range[1], result.Total))
	R.RenderJSON(w, result)
}

func (g *groupHandlers) groupUpdateCtrl(w http.ResponseWriter, r *http.Request) { //nolint dupl
	groupId := chi.URLParam(r, "id")

	i, err := strconv.ParseInt(groupId, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, g.l, http.StatusBadRequest, err, "failed to parse group id with api")
		return
	}

	group, err := g.dataStore.GetGroup(r.Context(), i)
	if err != nil {
		SendErrorJSON(w, r, g.l, http.StatusInternalServerError, err, "failed to get group with api")
		return
	}

	if err = json.NewDecoder(r.Body).Decode(&group); err != nil {
		SendErrorJSON(w, r, g.l, http.StatusInternalServerError, err, "failed to decode group data for update with api")
		return
	}

	if err = g.dataStore.UpdateGroup(r.Context(), group); err != nil {
		SendErrorJSON(w, r, g.l, http.StatusInternalServerError, err, "failed to update group data with api")
		return
	}

	R.RenderJSON(w, responseMessage{
		Error: false,
		ID:    group.ID,
		Data:  group,
	})
}

func (g *groupHandlers) groupDeleteCtrl(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		SendErrorJSON(w, r, g.l, http.StatusBadRequest, err, "failed to parse group id with api")
		return
	}

	if err := g.dataStore.DeleteGroup(r.Context(), id); err != nil {
		SendErrorJSON(w, r, g.l, http.StatusInternalServerError, err, "failed to delete group with api")
		return
	}

	R.RenderJSON(w, responseMessage{Message: "group deleted"})
}
