package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/distribution/notifications"
	"github.com/go-pkgz/auth/token"
	"github.com/go-pkgz/rest"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// dataServiceInterface implement dataService instance
type dataServiceInterface interface {
	RepositoriesMaintenance(ctx context.Context, timeout int64)
	RepositoryEventsProcessing(ctx context.Context, envelope notifications.Envelope) (err error)
	SyncExistedRepositories(ctx context.Context) error
}

// registryHandlers implement controllers which allow manipulation with registry entries using REST API endpoints
type registryHandlers struct {
	endpointsHandler
	registryService registryInterface
	dataService     dataServiceInterface
}

func (rh *registryHandlers) tokenAuth(w http.ResponseWriter, r *http.Request) {

	username, password, ok := r.BasicAuth()
	if !ok || password == "" {
		// check access for all (guest), public repository access
		// owner_id equal '0' for anonymous user
		user := store.User{ID: engine.AnonymousUserID}
		rh.parseTokenRequestParams(w, r, user)
		return
	}

	user, errUser := rh.dataStore.GetUser(r.Context(), username)
	if errUser != nil {
		anonUser := store.User{ID: engine.AnonymousUserID}
		rh.parseTokenRequestParams(w, r, anonUser)
		return
	}

	if !store.ComparePassword(user.Password, password) || user.Disabled {
		rh.l.Logf("wrong user credentials or account disabled: %s", user.Login)
		renderJSONWithStatus(
			w,
			registryResponseError(registry.APIError{Code: "DENIED", Message: "access denied"}),
			http.StatusForbidden,
		)
		return
	}

	rh.parseTokenRequestParams(w, r, user)
}

// health checks availability a registry service
func (rh *registryHandlers) health(w http.ResponseWriter, r *http.Request) {

	if err := rh.registryService.APIVersionCheck(r.Context()); err != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "registry service request failed")
		return
	}
	rest.RenderJSON(w, responseMessage{Message: "ok"})
}

// events handle registry event and pass them to a relevant service which will processing an events.
// In particular this service should contain information about all repositories in registry, but registry hasn't API
// for get repository by name and return repository entries with set up to 100 items per each request
// and more with cursor pagination.
// This handler catch events from repository, extract repository data from one and store it in a storage of service.
func (rh *registryHandlers) events(w http.ResponseWriter, r *http.Request) {

	var eventsEnvelope notifications.Envelope

	if err := json.NewDecoder(r.Body).Decode(&eventsEnvelope); err != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "failed to parse notification event")
		return
	}
	defer func() { _ = r.Body.Close() }()

	if err := rh.dataService.RepositoryEventsProcessing(r.Context(), eventsEnvelope); err != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "failed to processing event message from registry")
	}
	rest.RenderJSON(w, responseMessage{Message: "ok"})
}

func (rh *registryHandlers) imageConfig(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query()["name"]
	digest := r.URL.Query()["digest"]
	if len(name) < 1 || len(digest) < 1 {
		err := fmt.Errorf("params name and digest must be set")
		SendErrorJSON(w, r, rh.l, http.StatusBadRequest, err, err.Error())
		return
	}
	blob, err := rh.registryService.GetBlob(r.Context(), name[0], digest[0])
	if err != nil {
		err = fmt.Errorf("failed to retrieve blobs data for repo: %s digest: %s err: %v", name, digest, err)
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "failed to retrieve blobs data")
		return
	}
	rest.RenderJSON(w, responseMessage{Data: map[string]interface{}{"id": 0, "value": blob}, Message: "ok"})
}

func (rh *registryHandlers) deleteDigest(w http.ResponseWriter, r *http.Request) {
	digest := r.URL.Query()["digest"]
	name := r.URL.Query()["name"]

	if len(name) < 1 || len(digest) < 1 {
		err := fmt.Errorf("params name and digest must be set")
		SendErrorJSON(w, r, rh.l, http.StatusBadRequest, err, err.Error())
		return
	}

	if err := rh.registryService.DeleteTag(r.Context(), name[0], digest[0]); err != nil {
		rh.l.Logf("%v", err)
		err = fmt.Errorf("delete digest fail: %v", err)
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, err.Error())
	}
}

// syncRepositories runs task for check existed entries in a registry service and synchronize it with storage
func (rh *registryHandlers) syncRepositories(w http.ResponseWriter, r *http.Request) {
	if err := rh.dataService.SyncExistedRepositories(rh.ctx); err != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "failed to run repositories sync task")
		return
	}
	rest.RenderJSON(w, responseMessage{Message: "ok", Data: []interface{}{}})
}

// catalogList returns list of repositories entry
func (rh *registryHandlers) catalogList(w http.ResponseWriter, r *http.Request) {
	filter, err := engine.FilterFromURLExtractor(r.URL)
	if err != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "failed to parse URL parameters for make query filter")
		return
	}

	groupBy, isGroupBy := r.URL.Query()["group_by"]

	user, err := token.GetUserInfo(r)
	if err != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, err, "failed to get current user")
		return
	}

	if user.GetRole() == store.UserRole {
		if filter.Filters == nil {
			filter.Filters = map[string]interface{}{engine.RepositoriesByUserAccess: user.Attributes["uid"]}
		} else {
			filter.Filters[engine.RepositoriesByUserAccess] = user.Attributes["uid"]
		}
	}

	filter.GroupByField = !isGroupBy || (len(groupBy) > 0 && groupBy[0] != "none")
	repoList, errReposList := rh.dataStore.FindRepositories(r.Context(), filter)

	if errReposList != nil {
		SendErrorJSON(w, r, rh.l, http.StatusInternalServerError, errReposList, "failed to fetch list of repositories")
		return
	}
	w.Header().Add("Content-Range", fmt.Sprintf("registry/catalog %d-%d/%d", filter.Range[0], filter.Range[1], repoList.Total))
	rest.RenderJSON(w, repoList)

}

func (rh *registryHandlers) parseTokenRequestParams(w http.ResponseWriter, r *http.Request, user store.User) {
	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		renderJSONWithStatus(
			w,
			registryResponseError(registry.APIError{
				Code:    "UNSUPPORTED",
				Message: " failed to parse request params",
			}),
			http.StatusBadRequest)
		return
	}

	// processing push and pull requests
	if queryParams.Get("scope") != "" {
		scopeParts := strings.Split(queryParams.Get("scope"), ":")

		// skips account scope check for anonymous users only
		if len(scopeParts) != 3 || (user.ID != engine.AnonymousUserID && queryParams.Get("account") != user.Login) {
			rh.l.Logf("[ERROR] wrong scope or user and account doesn't match: %s :user %s", r.RequestURI, user.Login)

			renderJSONWithStatus(
				w,
				registryResponseError(registry.APIError{
					Code:    "UNSUPPORTED",
					Message: fmt.Sprintf("wrong scope or user and account doesn't match: %s :user %s", r.RequestURI, user.Login),
				}),
				http.StatusBadRequest)
			return
		}

		tokenRequest := registry.TokenRequest{
			Account: queryParams.Get("account"),
			Service: queryParams.Get("service"),
			Type:    scopeParts[0],
			Name:    scopeParts[1],
			Actions: strings.Split(scopeParts[2], ","),
		}

		if expireTime := queryParams.Get("expire"); expireTime != "" {
			expireValue, errExpireConvert := strconv.ParseInt(expireTime, 10, 64)
			if errExpireConvert != nil {
				errValue := fmt.Errorf("expire value must be a number: %v", errExpireConvert)
				rh.l.Logf("[ERROR] %v", errValue)
				renderJSONWithStatus(
					w,
					registryResponseError(registry.APIError{Code: "UNSUPPORTED", Message: errValue.Error()}),
					http.StatusBadRequest)
				return
			}
			tokenRequest.ExpireTime = expireValue
		}

		if allow, errCheck := rh.checkUserAccess(r.Context(), user, tokenRequest); !allow || errCheck != nil {
			errMsg := fmt.Errorf("[ERROR] access to registry resource not allowed for user %s: %v", user.Login, errCheck)

			rh.l.Logf("%v", errMsg)
			renderJSONWithStatus(
				w,
				registryResponseError(registry.APIError{Code: "", Message: errMsg.Error()}),
				http.StatusForbidden)
			return
		}

		tokenString, errToken := rh.registryService.Token(tokenRequest)
		if errToken != nil {
			rh.l.Logf("[ERROR] failed to issue token for request: %s", r.RequestURI)
			renderJSONWithStatus(
				w,
				registryResponseError(registry.APIError{Code: "UNSUPPORTED", Message: "access denied"}),
				http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if _, err = w.Write([]byte(tokenString)); err != nil {
			rh.l.Logf("[ERROR] failed to response for token request: %v", err)
		}
		return
	}

	// processing docker login requests
	if queryParams.Get("account") != "" && queryParams.Get("client_id") != "" && user.ID != engine.AnonymousUserID {
		userToken, errLogin := rh.registryService.Login(user)
		if errLogin != nil || queryParams.Get("account") != user.Login {
			rh.l.Logf("[ERROR] failed to processing docker login request: %v", errLogin)
			w.WriteHeader(http.StatusInternalServerError)
			renderJSONWithStatus(
				w,
				registryResponseError(registry.APIError{Code: "UNAUTHORIZED", Message: "access denied", Detail: "access not allowed for user"}),
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if _, err = w.Write([]byte(userToken)); err != nil {
			rh.l.Logf("[ERROR] failed to response for docker login: %v", err)
		}
		return
	}

	renderJSONWithStatus(
		w,
		registryResponseError(registry.APIError{Code: "UNAUTHORIZED", Message: "access denied", Detail: "bad request for resource access"}),
		http.StatusForbidden,
	)

}

func (rh *registryHandlers) checkUserAccess(ctx context.Context, user store.User, tokenRequest registry.TokenRequest) (bool, error) {

	if user.Role == "admin" {
		return true, nil
	}

	filter := engine.QueryFilter{
		Filters: map[string]interface{}{
			"owner_id":      user.ID,
			"resource_type": tokenRequest.Type,
			"resource_name": tokenRequest.Name,
			"action":        tokenRequest.Actions,
			"disabled":      false,
		},
	}
	access, err := rh.dataStore.FindAccesses(ctx, filter)
	if err != nil {
		// check access to repository with for all users permission
		if !errors.Is(err, engine.ErrNotFound) {
			return false, err
		}
		if user.ID > 0 {
			filter.Filters["owner_id"] = engine.RegisteredUserID
			if access, err = rh.dataStore.FindAccesses(ctx, filter); access.Total > 0 {
				return true, err
			}
		}

		filter.Filters["owner_id"] = engine.AnonymousUserID
		if access, err = rh.dataStore.FindAccesses(ctx, filter); err != nil {
			return false, err
		}
	}

	// if at least one item exist it's mean that access for user exist
	return access.Total > 0, nil
}

// registryErrors when registry response is failure, covered in detail in their relevant sections, are reported as part of 4xx responses, in a json response body.
// One or more errors will be returned in this format
type registryErrors struct {
	Errors []registry.APIError `json:"errors"`
}

func registryResponseError(err ...registry.APIError) (errs registryErrors) {
	errs.Errors = append(errs.Errors, err...)
	return errs
}
