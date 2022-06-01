package registry

// This is mock implementation of docker registry V2 api for use in unit tests

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type handlerFn func(w http.ResponseWriter, r *http.Request)

// MockRegistry represent a registry mock
type MockRegistry struct {
	server   *httptest.Server
	hostPort string
	handlers map[string]handlerFn
	t        testing.TB
	mux      http.ServeMux
	mu       sync.Mutex
}

// RegisterHandler register the specified handler for the registry mock
func (mr *MockRegistry) RegisterHandler(path string, h handlerFn) {
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.handlers[path] = h
}

// NewMockRegistry creates a registry mock
func NewMockRegistry(t testing.TB, host string, port int) *MockRegistry {
	t.Helper()
	testRegistry := &MockRegistry{handlers: make(map[string]handlerFn)}
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
}

func (mr *MockRegistry) apiVersionCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.Header().Set("docker-distribution-api-version", "registry/2.0")
	_, err := w.Write([]byte("{}"))
	assert.NoError(mr.t, err)
}
