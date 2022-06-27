package server

import (
	"context"
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
