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
	var (
		isRepoFound bool
		isTagFound  bool
	)
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
	w.Header().Set("docker-content-digest", "sha256:5c3b3ba876c7e23bdf06f5657a57774420c38b290b9ffa5635cc70f7d68cb117")
	_, err := w.Write([]byte(testManifest))
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
