package registry

// This is mock implementation of docker registry V2 api for use in unit tests

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/libtrust"
	"github.com/golang-jwt/jwt"
	"github.com/zebox/registry-admin/app/store"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"testing"

	"github.com/go-pkgz/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultMockUsername = "test_admin"
	defaultMockPassword = "test_password"
)

// tokenProcessing is functions for  parse www-authenticate header and request jwt token with credentials for get access to registry resources based on token claims data
type tokenProcessing interface {
	Token(request *http.Request) (string, error)
	ParseAuthenticateHeaderRequest(wwwRequest string) (AuthorizationRequest, error)
}

type repositories struct {
	List []string `json:"repositories"`
}

type tags struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// MockRegistry represent a registry mock
type MockRegistry struct {
	server   *httptest.Server
	hostPort string
	handlers map[*regexp.Regexp]http.Handler
	repositories
	tagList []tags

	auth        authType
	credentials struct {
		username string
		password string
		access   store.Access
	}

	tokenFn   tokenProcessing
	publicKey libtrust.PublicKey

	t   testing.TB
	mux *http.ServeMux
}

type MockRegistryOptions func(option *MockRegistry)

func TokenAuth(tokenFn tokenProcessing) MockRegistryOptions {
	return func(mr *MockRegistry) {
		mr.auth = SelfToken
		mr.tokenFn = tokenFn
	}
}

func Credentials(username, password string, access store.Access) MockRegistryOptions {
	if username == "" {
		username = defaultMockUsername
	}
	return func(mr *MockRegistry) {
		mr.credentials.username = username
		mr.credentials.password = password
		mr.credentials.access = access
	}
}

func PublicKey(publicKey libtrust.PublicKey) MockRegistryOptions {
	return func(mr *MockRegistry) {
		mr.publicKey = publicKey
	}
}

// NewMockRegistry creates a registry mock
func NewMockRegistry(t testing.TB, host string, port int, repoNumber, tagNumber int, opts ...MockRegistryOptions) *MockRegistry {
	t.Helper()
	testRegistry := &MockRegistry{handlers: make(map[*regexp.Regexp]http.Handler)}
	testRegistry.t = t
	testRegistry.prepareRepositoriesData(repoNumber, tagNumber)
	testRegistry.prepareRegistryMockEndpoints()
	testRegistry.mux = http.NewServeMux()

	// set default credentials for basic auth
	testRegistry.credentials.username = defaultMockUsername
	testRegistry.credentials.password = defaultMockPassword

	for _, opt := range opts {
		opt(testRegistry)
	}

	testRegistry.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !testRegistry.authCheck(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		for k, v := range testRegistry.handlers {
			if k.MatchString(r.URL.Path) {

				// Delete method has only one handler
				if r.Method == "DELETE" {
					http.HandlerFunc(testRegistry.deleteManifest).ServeHTTP(w, r)
					return
				}
				v.ServeHTTP(w, r)
				return
			}
		}

	})

	// prepare test http server
	testRegistry.hostPort = fmt.Sprintf("%s:%d", host, port)
	l, err := net.Listen("tcp", testRegistry.hostPort)
	require.Nil(t, err)
	ts := httptest.NewUnstartedServer(testRegistry.mux)
	assert.NoError(t, ts.Listener.Close())
	ts.Listener = l
	testRegistry.server = ts
	ts.Start()

	return testRegistry
}

// URL returns the url of the registry
func (mr *MockRegistry) URL() string {
	return fmt.Sprintf("http://%s", mr.hostPort)
}

// Close closes mock and releases resources
func (mr *MockRegistry) Close() {
	mr.server.Close()
}

func (mr *MockRegistry) prepareRegistryMockEndpoints() {
	// Api Version Check
	if mr.handlers == nil {
		mr.t.Fatal("failed to prepare handler while handlers undefined")
	}

	// bind tests docker registry api endpoints
	mr.handlers[regexp.MustCompile(`/v2/$`)] = http.HandlerFunc(mr.apiVersionCheck)
	mr.handlers[regexp.MustCompile(`/v2/_catalog+`)] = http.HandlerFunc(mr.getCatalog)
	mr.handlers[regexp.MustCompile(`/v2/(.*)/tags/+`)] = http.HandlerFunc(mr.getImageTags)
	mr.handlers[regexp.MustCompile(`/v2/(.*)/manifests/(.*)`)] = http.HandlerFunc(mr.getManifest)

}

func (mr *MockRegistry) prepareRepositoriesData(repoNumbers, tagNumbers int) {
	if repoNumbers < 1 {
		repoNumbers = 50
	}

	if tagNumbers < 1 {
		tagNumbers = 50
	}

	var testRepos []string
	// filling repository list
	for i := 0; i < repoNumbers; i++ {
		var testTags []string
		testRepoName := fmt.Sprintf("test_repo_%d", i)
		testRepos = append(testRepos, testRepoName)

		// filling tag list
		for j := 0; j < tagNumbers; j++ {
			testTagName := fmt.Sprintf("test_tag_%d", j)
			testTags = append(testTags, testTagName)
		}

		mr.tagList = append(mr.tagList, tags{
			Name: testRepoName,
			Tags: testTags,
		})
	}
	mr.repositories.List = testRepos
}

func (mr *MockRegistry) authCheck(req *http.Request) bool {
	switch mr.auth {
	case Basic:
		if username, passwd, ok := req.BasicAuth(); ok {
			return username == mr.credentials.username && passwd == mr.credentials.password
		}
	case SelfToken:

		username, passwd, ok := req.BasicAuth()
		if !ok {
			return ok
		}
		if mr.credentials.username != username || mr.credentials.password != passwd {
			return false
		}

		var repoNameRE = regexp.MustCompile(`/v2/(.*)/tags`)
		repoName := repoNameRE.FindStringSubmatch(req.URL.Path)
		if len(repoName) == 0 {
			return false
		}

		headerValue := fmt.Sprintf(`Bearer realm="http://127.0.0.1/token",service="127.0.0.1",scope="repository:%s:*"`, repoName[1])
		authRequest, errAuth := mr.tokenFn.ParseAuthenticateHeaderRequest(headerValue)
		require.NoError(mr.t, errAuth)
		req.Header.Set(authenticateHeaderName, headerValue)
		if mr.credentials.access.ResourceName != authRequest.Name || mr.credentials.access.Disabled {
			return false
		}

		token, err := mr.tokenFn.Token(req)
		require.NoError(mr.t, err)

		var authToken clientToken
		err = json.Unmarshal([]byte(token), &authToken)
		require.NoError(mr.t, err)

		_, claims, err := mr.parseHeaderForJwt(authToken.Token)
		if err != nil {
			return false
		}
		accessData := claims["access"].([]interface{})
		accessClaims := accessData[0].(map[string]interface{})
		name := accessClaims["name"].(string)
		resourceType := accessClaims["type"].(string)
		actions := accessClaims["actions"].([]interface{})

		return name == repoName[1] && resourceType == "repository" && func(actions []interface{}) bool {
			for _, a := range actions {
				if a.(string) == "*" {
					return true
				}
			}
			return false
		}(actions)
	}

	return false
}

func (mr *MockRegistry) parseHeaderForJwt(authToken string) (*jwt.Token, jwt.MapClaims, error) {

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(authToken, claims, func(token *jwt.Token) (interface{}, error) {
		if mr.publicKey == nil {
			return nil, errors.New("wrong public key")
		}
		return mr.publicKey.CryptoPublicKey(), nil
	})
	return token, claims, err

}
func (mr *MockRegistry) apiVersionCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.Header().Set("docker-distribution-api-version", "registry/2.0")
	_, err := w.Write([]byte("{}"))
	assert.NoError(mr.t, err)
}

func (mr *MockRegistry) getCatalog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.Header().Set("docker-distribution-api-version", "registry/2.0")
	urlFragments, err := url.ParseQuery(r.URL.RawQuery)
	assert.NoError(mr.t, err)
	if urlFragments.Get("n") == "" {
		data, err := json.Marshal(mr.repositories)
		assert.NoError(mr.t, err)
		_, err = w.Write(data)
		assert.NoError(mr.t, err)
		return
	}

	n := urlFragments.Get("n")
	last := urlFragments.Get("last")
	isNext, lastIndex, result, errPagination := mr.preparePaginationResult(mr.repositories.List, n, last)
	require.NoError(mr.t, errPagination)
	rel := ` rel="next"`
	if !isNext {
		rel = ""
	}

	data, err := json.Marshal(Repositories{List: result})
	assert.NoError(mr.t, err)

	nextLinkUrl := fmt.Sprintf("/v2/_catalog?last=%s&n=%s; %s", mr.repositories.List[lastIndex], n, rel)
	w.Header().Set("link", nextLinkUrl)
	_, err = w.Write(data)
	assert.NoError(mr.t, err)

}

func (mr *MockRegistry) getImageTags(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.Header().Set("docker-distribution-api-version", "registry/2.0")

	// extractRepoName
	var repoNameRE = regexp.MustCompile(`(?m)/v2/(.*)/tags/`)
	repoName := repoNameRE.FindStringSubmatch(r.URL.Path)
	if len(repoName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var (
		tags        []string
		isRepoFound bool
		repoIndex   int
	)

	for i, v := range mr.tagList {
		if v.Name == repoName[1] {
			tags = v.Tags
			isRepoFound = true
			repoIndex = i
			break
		}

	}

	if !isRepoFound {
		apiError := ApiError{
			Code:    "NAME_UNKNOWN",
			Message: "repository name not known to registry",
		}
		apiError.Detail = map[string]string{"name": repoName[1]}
		w.WriteHeader(http.StatusNotFound)
		rest.RenderJSON(w, apiError)
		return
	}

	urlFragments, err := url.ParseQuery(r.URL.RawQuery)
	assert.NoError(mr.t, err)
	if urlFragments.Get("n") == "" {
		data, err := json.Marshal(ImageTags{Name: repoName[1], Tags: tags})
		assert.NoError(mr.t, err)
		_, err = w.Write(data)
		assert.NoError(mr.t, err)
		return
	}

	n := urlFragments.Get("n")
	last := urlFragments.Get("last")
	isNext, lastIndex, result, errPagination := mr.preparePaginationResult(tags, n, last)
	require.NoError(mr.t, errPagination)
	rel := ` rel="next"`
	if !isNext {
		rel = ""
	}

	data, err := json.Marshal(ImageTags{Name: repoName[1], Tags: result})
	assert.NoError(mr.t, err)

	nextLinkUrl := fmt.Sprintf("/v2/_catalog?last=%s&n=%s; %s", mr.tagList[repoIndex].Tags[lastIndex], n, rel)
	w.Header().Set("link", nextLinkUrl)
	_, err = w.Write(data)
	assert.NoError(mr.t, err)

}

func (mr *MockRegistry) getManifest(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	testManifest := `{
    "schemaVersion": 2,
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "config": {
        "mediaType": "application/vnd.docker.container.image.v1+json",
        "size": 4120,
        "digest": "sha256:f03b14782dfdb2d0e1331c19161fac0f09e7dcd294116e06dd2c50acd041f1db"
    },
    "layers": [
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 2811478,
            "digest": "sha256:5843afab387455b37944e709ee8c78d7520df80f8d01cf7f861aae63beeddb6b"
        },
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 92,
            "digest": "sha256:1d9a043fcb62927e88cb939d7776a0f776d127c390177343b621023a549e4eff"
        },
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 10142128,
            "digest": "sha256:0bcf3b0e371981104ecf6ce7e17c52775e18e7897d9193b52550a48772bc0047"
        },
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 10142152,
            "digest": "sha256:560beb475c9a630a4cbb206d115a9b387ceda4348cd8d6cd3c386afae01d8775"
        },
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 1236,
            "digest": "sha256:74c6258afffc26bf57a056dee76077bd300ae64bdd6e8745d6dc41b78e3c5152"
        },
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 12340918,
            "digest": "sha256:51d2eca1d7720849b7f120514745529a31140683964be2f66bec72df4c430a81"
        },
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 115,
            "digest": "sha256:5b06090241844dbdb6a1e76387301a77657cbbc09dc1dbe5152046377d9c7c4e"
        },
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 114,
            "digest": "sha256:756d41630d29aa89226b5aecb4012836dd3ed4933dec45580eb04e2ce4ea228a"
        },
        {
            "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
            "size": 115,
            "digest": "sha256:c759b8c5c3b21b188bb1ad5c1e7bd4d63b73b73f85dca622f4d415422c283196"
        }
    ]
}
`
	// extractRepoName
	var repoNameRE = regexp.MustCompile(`/v2/(.*)/manifests/(.*)`)
	requestData := repoNameRE.FindStringSubmatch(r.URL.Path)
	if len(requestData) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// search for repo and tags
	var isRepoFound, isTagFound bool

	for _, v := range mr.tagList {
		if v.Name == requestData[1] {
			isRepoFound = true
			for _, tag := range v.Tags {
				if tag == requestData[2] {
					isTagFound = true
				}
			}
			break
		}
	}

	if !isRepoFound || !isTagFound {
		apiError := ApiError{
			Code:    "NAME_UNKNOWN",
			Message: "either repository name or tag not not found in registry",
		}
		detail := map[string]string{}
		if !isRepoFound {
			detail["repository"] = requestData[1]
		}
		if !isRepoFound {
			detail["tag"] = requestData[2]
		}
		apiError.Detail = detail
		w.WriteHeader(http.StatusNotFound)
		rest.RenderJSON(w, apiError)
		return
	}

	w.Header().Set("content-type", "application/vnd.docker.distribution.manifest.v2+json")
	w.Header().Set("docker-distribution-api-version", "registry/2.0")

	w.Header().Set("docker-content-digest", "sha256:"+makeDigest(requestData[2]))
	_, err := w.Write([]byte(testManifest))
	assert.NoError(mr.t, err)
}

func (mr *MockRegistry) deleteManifest(w http.ResponseWriter, r *http.Request) {

	var repoNameRE = regexp.MustCompile(`/v2/(.*)/manifests/(.*)`)
	requestData := repoNameRE.FindStringSubmatch(r.URL.Path)
	if len(requestData) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var isTagFound, isRepoFound bool

	for i, v := range mr.tagList {
		if v.Name == requestData[1] {
			isRepoFound = true
			for j, tag := range v.Tags {
				digest := makeDigest(tag)
				if digest == requestData[2] {
					isTagFound = true
					updatedTags := append(v.Tags[:j], v.Tags[j+1:]...)
					mr.tagList[i].Tags = updatedTags
					break
				}
			}
			break
		}
	}

	if !isRepoFound || !isTagFound {
		apiError := ApiError{
			Code:    "NAME_UNKNOWN",
			Message: "either repository name or tag not not found in registry",
		}
		detail := map[string]string{}
		if !isRepoFound {
			detail["repository"] = requestData[1]
		}
		if !isRepoFound {
			detail["tag"] = requestData[2]
		}
		apiError.Detail = detail
		w.WriteHeader(http.StatusNotFound)
		rest.RenderJSON(w, apiError)
		return
	}

}

func (mr *MockRegistry) preparePaginationResult(items []string, n, last string) (isNext bool, lastIndex int, result []string, err error) {
	// search last index
	pages, err := strconv.Atoi(n)
	if err != nil {
		return false, lastIndex, nil, err
	}

	if last != "" {
		for i, v := range items {
			if v == last {
				lastIndex = i
				break
			}
		}
	}

	next := lastIndex + pages

	if next < len(items) {
		result = items[lastIndex:next]
		isNext = true
		lastIndex = next
	}

	if !isNext {
		result = items[lastIndex:]
		lastIndex = len(items) - 1
	}
	return isNext, lastIndex, result, err
}

func makeDigest(data string) string {
	hasher := sha256.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}
