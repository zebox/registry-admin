package registry

// This is mock implementation of docker registry V2 api for use in unit tests

import (
	"encoding/json"
	"fmt"
	"github.com/go-pkgz/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"testing"
)

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

	t   testing.TB
	mux *http.ServeMux
}

// NewMockRegistry creates a registry mock
func NewMockRegistry(t testing.TB, host string, port int, repoNumber, tagNumber int) *MockRegistry {
	t.Helper()
	testRegistry := &MockRegistry{handlers: make(map[*regexp.Regexp]http.Handler)}
	testRegistry.prepareRepositoriesData(repoNumber, tagNumber)
	testRegistry.prepareRegistryMockEndpoints()
	testRegistry.mux = http.NewServeMux()

	testRegistry.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		for k, v := range testRegistry.handlers {
			if k.MatchString(r.URL.Path) {
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
	mr.handlers[regexp.MustCompile(`/v2/_catalog`)] = http.HandlerFunc(mr.getCatalog)
	mr.handlers[regexp.MustCompile(`/v2/(.*)/tags/+`)] = http.HandlerFunc(mr.getImageTags)

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
