package registry

// This is mock implementation of docker registry V2 api for use in unit tests

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync"
	"testing"
)

type handlerFn func(w http.ResponseWriter, r *http.Request)

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
	handlers map[string]handlerFn
	repositories
	tagList []tags

	t   testing.TB
	mux http.ServeMux
	mu  sync.Mutex
}

// RegisterHandler register the specified handler for the registry mock
func (mr *MockRegistry) RegisterHandler(path string, h handlerFn) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.handlers[path] = h
}

// NewMockRegistry creates a registry mock
func NewMockRegistry(t testing.TB, host string, port int, repoNumber, tagNumber int) *MockRegistry {
	t.Helper()
	testRegistry := &MockRegistry{handlers: make(map[string]handlerFn)}
	testRegistry.prepareRepositoriesData(repoNumber, tagNumber)
	testRegistry.prepareRegistryMockEndpoints()
	mux := http.NewServeMux()

	for k, v := range testRegistry.handlers {
		mux.Handle(k, http.HandlerFunc(v))
	}

	// prepare test http server
	testRegistry.hostPort = fmt.Sprintf("%s:%d", host, port)
	l, err := net.Listen("tcp", testRegistry.hostPort)
	require.Nil(t, err)
	ts := httptest.NewUnstartedServer(mux)
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
	mr.handlers["/v2/"] = mr.apiVersionCheck
	mr.handlers["/v2/_catalog"] = mr.getCatalog
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

func (mr *MockRegistry) apiVersionCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.Header().Set("docker-distribution-api-version", "registry/2.0")
	_, err := w.Write([]byte("{}"))
	assert.NoError(mr.t, err)
}

func (mr *MockRegistry) getCatalog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.Header().Set("docker-distribution-api-version", "registry/2.0")
	urlFragments, err := url.ParseQuery(r.URL.Fragment)
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
	pages, err := strconv.Atoi(n)
	require.NoError(mr.t, err)

	// search last index
	lastIndex := 0
	if last != "" {
		for i, v := range mr.repositories.List {
			if v == last {
				lastIndex = i
			}
		}
	}

	result := mr.repositories.List[lastIndex:]
	next := lastIndex + pages
	if (lastIndex + pages) < len(mr.repositories.List) {
		result = mr.repositories.List[lastIndex:next]
	}

	data, err := json.Marshal(result)
	assert.NoError(mr.t, err)

	lastRepo := mr.repositories.List[lastIndex]
	rel := ` rel="next"`
	if lastIndex == len(mr.repositories.List)-1 {
		rel = ""
	}
	nextLinkUrl := fmt.Sprintf("/v2/_catalog?last=%s&n=%d; %s", lastRepo, pages, rel)
	w.Header().Set(http.CanonicalHeaderKey("Link"), nextLinkUrl)
	_, err = w.Write(data)
	assert.NoError(mr.t, err)

}
