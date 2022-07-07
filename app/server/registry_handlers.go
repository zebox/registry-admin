package server

import (
	"context"
	"encoding/json"
	"github.com/docker/distribution/notifications"
	"github.com/go-pkgz/rest"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"net/http"
	"net/url"
	"strings"
)

// registryHandlers implement controllers which allow manipulation with registry entries using REST API endpoints
type registryHandlers struct {
	endpointsHandler
	registryService registryInterface
}

func (rh *registryHandlers) tokenAuth(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || password == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user, errUser := rh.dataStore.GetUser(r.Context(), username)
	if errUser != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !store.ComparePassword(user.Password, password) || user.Disabled {
		rh.l.Logf("wrong user credentials or account disabled: %s", user.Login)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// processing push and pull requests
	if queryParams.Get("scope") != "" {
		scopeParts := strings.Split(queryParams.Get("scope"), ":")
		if len(scopeParts) != 3 || queryParams.Get("account") != user.Login {
			rh.l.Logf("[ERROR] wrong scope or user and account doesn't match: %s :user %s", r.RequestURI, user.Login)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenRequest := registry.TokenRequest{
			Account: queryParams.Get("account"),
			Service: queryParams.Get("service"),
			Type:    scopeParts[0],
			Name:    scopeParts[1],
			Actions: strings.Split(scopeParts[2], ","),
		}

		if allow, errCheck := rh.checkUserAccess(r.Context(), user, tokenRequest); !allow || errCheck != nil {
			rh.l.Logf("[ERROR] access to registry resource not allowed for user %s: %v", user.Login, err)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		tokenString, errToken := rh.registryService.Token(tokenRequest)
		if errToken != nil {
			rh.l.Logf("[ERROR] failed to issue token for request: %s", r.RequestURI)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = w.Write([]byte(tokenString))
		return
	}

	// processing docker login requests
	if queryParams.Get("account") != "" && queryParams.Get("client_id") != "" {
		userToken, errLogin := rh.registryService.Login(user)
		if errLogin != nil || queryParams.Get("account") != user.Login {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = w.Write([]byte(userToken))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// health checks availability a registry service
func (rh *registryHandlers) health(w http.ResponseWriter, r *http.Request) {

	if err := rh.registryService.ApiVersionCheck(r.Context()); err != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "registry service request failed")
		return
	}
	rest.RenderJSON(w, responseMessage{Message: "ok"})
}

// events handle registry event and pass them to a relevant service which will processing an events.
// In particular this service should has information about all repositories in registry, but registry hasn't API for get repository by name
// and return repository entries with set up to 100 items per each request.
func (rh *registryHandlers) events(w http.ResponseWriter, r *http.Request) {

	var eventData notifications.Envelope

	if err := json.NewDecoder(r.Body).Decode(&eventData); err != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "failed to parse notification event")
		return
	}
	defer func() { _ = r.Body.Close() }()

	// rh.registryService.DeleteTag()
	/*digest := eventData.Events[0].Target.Descriptor.Digest
	//tag := eventData.Events[0].Target.Tag

	repo := eventData.Events[0].Target.Repository
	if err := rh.registryService.DeleteTag(r.Context(), repo, digest.String()); err != nil {
		rh.l.Logf("%v", err)
	}*/
	rh.l.Logf("%s", eventData.Events[0].Action)
	rest.RenderJSON(w, responseMessage{Message: "ok"})
}

func (rh *registryHandlers) delete(w http.ResponseWriter, r *http.Request) {
	t := r.URL.Query()["tag"]
	n := r.URL.Query()["name"]
	// rh.registryService.DeleteTag()
	/*digest := eventData.Events[0].Target.Descriptor.Digest
	//tag := eventData.Events[0].Target.Tag

	repo := eventData.Events[0].Target.Repository*/
	if err := rh.registryService.DeleteTag(r.Context(), n[0], t[0]); err != nil {
		rh.l.Logf("%v", err)
	}
}

// catalogList returns list of repositories entry
func (rh *registryHandlers) catalogList(w http.ResponseWriter, r *http.Request) {
	/*filter, err := engine.FilterFromUrlExtractor(r.URL)
	if err != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "failed to parse URL parameters for make query filter")
		return
	}*/

	repoList, err := rh.registryService.Catalog(r.Context(), "", "")
	if err != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "registry service request failed")
		return
	}
	rest.RenderJSON(w, responseMessage{Message: "ok", Data: repoList})
}

func (rh *registryHandlers) checkUserAccess(ctx context.Context, user store.User, tokenRequest registry.TokenRequest) (bool, error) {

	if user.Role == "admin" {
		return true, nil
	}

	filter := engine.QueryFilter{
		IDs: []int64{user.ID},
		Filters: map[string]interface{}{
			"owner_id":      user.ID,
			"resource_type": tokenRequest.Type,
			"resource_name": tokenRequest.Name,
			"action":        tokenRequest.Actions,
		},
	}
	access, err := rh.dataStore.FindAccesses(ctx, filter)
	if err != nil {
		return false, err
	}
	// if at least one item exist it's mean that access for user exist
	return access.Total > 0, nil
}
