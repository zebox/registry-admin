package server

// httpErrors is helper for render http errors with logging and misc parameters
// this idea borrow from package https://github.com/go-pkgz/rest and extended for use in this project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-pkgz/rest/logger"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

// SendErrorJSON sends {error: msg} with error code and logging error and caller
func SendErrorJSON(w http.ResponseWriter, r *http.Request, l logger.Backend, code int, err error, msg string) {
	if l != nil {
		l.Logf("%s", errDetailsMsg(r, code, err, msg))
	}
	errorResponse := responseMessage{
		Error:   true,
		Message: fmt.Sprintf("%s: %s", err, msg),
	}
	renderJSONWithStatus(w, errorResponse, code)
}

// renderJSONWithStatus sends data as json and enforces status code
func renderJSONWithStatus(w http.ResponseWriter, data interface{}, code int) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_, _ = w.Write(buf.Bytes())
}

func errDetailsMsg(r *http.Request, code int, err error, msg string) string {

	q := r.URL.String()
	if qun, e := url.QueryUnescape(q); e == nil {
		q = qun
	}

	srcFileInfo := ""
	if pc, file, line, ok := runtime.Caller(2); ok {
		fnameElems := strings.Split(file, "/")
		funcNameElems := strings.Split(runtime.FuncForPC(pc).Name(), "/")
		srcFileInfo = fmt.Sprintf(" [caused by %s:%d %s]", strings.Join(fnameElems[len(fnameElems)-3:], "/"),
			line, funcNameElems[len(funcNameElems)-1])
	}

	remoteIP := r.RemoteAddr
	if pos := strings.Index(remoteIP, ":"); pos >= 0 {
		remoteIP = remoteIP[:pos]
	}
	if err == nil {
		err = errors.New("no error")
	}
	return fmt.Sprintf("%s - %v - %d - %s - %s%s", msg, err, code, remoteIP, q, srcFileInfo)
}
