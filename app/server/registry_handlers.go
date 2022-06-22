package server

import (
	"github.com/zebox/registry-admin/app/store"
	"net/http"
)

// registryHandlers implement controllers which allow manipulation with registry entries using REST API endpoints
type registryHandlers struct {
	endpointsHandler
	registryService registryInterface
}

func (rh *registryHandlers) tokenAuth(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	user, errUser := rh.dataStore.GetUser(r.Context(), username)
	if errUser != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// login attempt if password not empty
	// GET /api/v1/registry/auth?account=admin&client_id=docker&offline_token=true&service=container_registry
	if password != "" {
		if !store.ComparePassword(user.Password, password) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		token, err := rh.registryService.Login(user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err = w.Write([]byte(token)); err != nil {
			rh.l.Logf("failed to write response for auth request: %v", err)
		}
		return
	}

	rh.l.Logf("%s %s", username, password)
}
