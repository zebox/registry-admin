package server

// this part of code borrow from https://github.com/umputun/reproxy/blob/master/app/proxy/ssl.go

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	log "github.com/go-pkgz/lgr"
	"golang.org/x/crypto/acme/autocert"

	R "github.com/go-pkgz/rest"
)

// sslMode defines ssl mode for rest server
type sslMode int8

const (
	// SSLNone defines to run http server only
	SSLNone sslMode = iota

	// SSLStatic defines to run both https and http server. Redirect http to https
	SSLStatic

	// SSLAuto defines to run both https and http server. Redirect http to https. Https server with autocert support
	SSLAuto
)

// SSLConfig holds all ssl params for rest server
type SSLConfig struct {
	SSLMode       sslMode
	Cert          string
	Key           string
	ACMELocation  string
	ACMEEmail     string
	FQDNs         []string
	Port          int // allow user define custom port for secure connection
	RedirHTTPPort int
}

// httpToHTTPSRouter creates new router which does redirect from http to https server
// with default middlewares. Used in 'static' ssl mode.
func (s *Server) httpToHTTPSRouter() http.Handler {
	log.Printf("[DEBUG] create http-to-https redirect routes")
	return R.Wrap(s.redirectHandler(), R.Recoverer(log.Default()))
}

// httpChallengeRouter creates new router which performs ACME "http-01" challenge response
// with default middlewares. This part is necessary to obtain certificate from LE.
// If it receives not an acme challenge it performs redirect to https server.
// Used in 'auto' ssl mode.
func (s *Server) httpChallengeRouter(m *autocert.Manager) http.Handler {
	log.Printf("[DEBUG] create http-challenge routes")
	return R.Wrap(m.HTTPHandler(s.redirectHandler()), R.Recoverer(log.Default()))
}

func (s *Server) redirectHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server := strings.Split(r.Host, ":")[0]
		newURL := fmt.Sprintf("https://%s:%d%s", server, s.SSLConfig.Port, r.URL.Path)
		if r.URL.RawQuery != "" {
			newURL += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, newURL, http.StatusTemporaryRedirect)
	})
}

func (s *Server) makeAutocertManager() *autocert.Manager {
	log.Printf("[DEBUG] autocert manager for domains: %+v, location: %s, email: %q",
		s.SSLConfig.FQDNs, s.SSLConfig.ACMELocation, s.SSLConfig.ACMEEmail)
	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(s.SSLConfig.ACMELocation),
		HostPolicy: autocert.HostWhitelist(s.SSLConfig.FQDNs...),
		Email:      s.SSLConfig.ACMEEmail,
	}
}

// makeHTTPSAutoCertServer makes https server with autocert mode (LE support)
func (s *Server) makeHTTPSAutocertServer(address string, router http.Handler, m *autocert.Manager) *http.Server {
	server := s.makeHTTPServer(address, router)
	cfg := s.makeTLSConfig()
	cfg.GetCertificate = m.GetCertificate
	server.TLSConfig = cfg
	return server
}

// makeHTTPSServer makes https server for static mode
func (s *Server) makeHTTPSServer(address string, router http.Handler) *http.Server {
	server := s.makeHTTPServer(address, router)
	server.TLSConfig = s.makeTLSConfig()
	return server
}

func (s *Server) makeTLSConfig() *tls.Config {
	return &tls.Config{
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519,
			tls.CurveP384,
		},
	}
}
