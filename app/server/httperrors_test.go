package server

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockLgr struct {
	buf bytes.Buffer
}

func (m *mockLgr) Logf(format string, args ...interface{}) {
	_, _ = m.buf.WriteString(fmt.Sprintf(format, args...))
}

func TestSendErrorJSON(t *testing.T) {
	l := &mockLgr{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/error" {
			t.Log("http err request", r.URL)
			SendErrorJSON(w, r, l, 500, errors.New("error 500"), "error details 123456")
			return
		}
		w.WriteHeader(404)
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/error")
	require.Nil(t, err)
	defer func() { assert.NoError(t, resp.Body.Close()) }()

	body, err := io.ReadAll(resp.Body)
	require.Nil(t, err)
	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "application/json; charset=utf-8", resp.Header.Get("content-type"))
	assert.Equal(t, `{"error":true,"message":"error 500: error details 123456","id":0,"data":null}`+"\n", string(body))
	t.Log(l.buf.String())
}

func TestErrorDetailsMsg(t *testing.T) {
	callerFn := func() {
		req, err := http.NewRequest("GET", "https://example.com/test?k1=v1&k2=v2", http.NoBody)
		require.Nil(t, err)
		req.RemoteAddr = "1.2.3.4"
		msg := errDetailsMsg(req, 500, errors.New("error 500"), "error details 123456")
		assert.Contains(t, msg, "error details 123456 - error 500 - 500 - 1.2.3.4 - https://example."+
			"com/test?k1=v1&k2=v2 [caused by")
		assert.Contains(t, msg, "app/server/httperrors_test.go:58 server.TestErrorDetailsMsg]", msg)

	}
	callerFn()
}

func TestErrorDetailsMsgNoError(t *testing.T) {
	callerFn := func() {
		req, err := http.NewRequest("GET", "https://example.com/test?k1=v1&k2=v2", http.NoBody)
		require.Nil(t, err)
		req.RemoteAddr = "1.2.3.4"
		msg := errDetailsMsg(req, 500, nil, "error details 123456")
		assert.Contains(t, msg, "error details 123456 - no error - 500 - 1.2.3.4 - https://example.com/test?k1=v1&k2=v2 [caused by")
		assert.Contains(t, msg, "app/server/httperrors_test.go:70 server.TestErrorDetailsMsgNoError]", msg)
	}
	callerFn()
}
