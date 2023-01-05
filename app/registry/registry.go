package registry

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// This is package implement features for interacts with instances of the docker registry,
// which is a service to manage information about docker images and enable their distribution using HTTP API V2 protocol
// detailed protocol description: https://docs.docker.com/registry/spec/api

const (
	// scheme version of manifest file
	// for details about scheme version goto https://docs.docker.com/registry/spec/manifest-v2-2/
	manifestSchemeV2 = "application/vnd.docker.distribution.manifest.v2+json"

	//  It uniquely identifies content by taking a collision-resistant hash of the bytes.
	contentDigestHeader = "docker-content-digest"
)

// authType define auth mechanism for accessing to docker registry using a docker HTTP API protocol
type authType int8

// UsersFn uses in adapter for bind FindUsers func in store engine with registry instance
type UsersFn func(ctx context.Context, filter engine.QueryFilter, withPassword bool) (users engine.ListResponse, err error)

const (
	// Basic allow access using auth basic credentials
	Basic authType = iota

	// SelfToken define this service as main auth/authz server for docker registry host
	SelfToken
)

var (
	// ErrNoMorePages used for cursor pagination state of registry entries
	ErrNoMorePages = errors.New("no more pages")
)

// Settings main configuration options for communicate with registry instance
type Settings struct {

	// Host is a fqdn of docker registry host
	// also it's value appends Subject Alternative Name for requested IP and Domain to certificate
	Host string

	// Port which registry accept requests
	Port uint

	// define authenticate type for access to docker registry api
	AuthType authType

	// use with basic auth only for dynamic update .htpasswd file
	HtpasswdPath string

	// credentials define user and login pair for auth in docker registry, when auth type set as basic
	credentials struct {
		login, password string
	}

	// The name of the service which hosts the resource.
	Service string

	// The name of the token issuer which hosts the resource.
	Issuer string

	// CertificatesPaths define a path to private, public keys and CA certificate.
	// If CertificatesPaths has all fields are empty, AccessToken will create keys by default, with default path.
	// If CertificatesPaths has all fields are empty, but certificates files exist AccessToken try to load existed keys and CA file.
	CertificatesPaths Certs

	// HttpsCert used when a https access cert and a token cert is differs. In this case HTTPSCert will add to trusted CA pool
	HTTPSCert string

	// InsecureRequest define option secure for make a https request to docker registry host, false by default
	InsecureRequest bool
}

// Registry is main instance for manipulation access of self-hosted docker registry
type Registry struct {
	settings Settings

	// use with basic auth only for dynamic update .htpasswd file
	htpasswd *htpasswd

	// use when auth with token is set
	registryToken *AccessToken

	httpClient *http.Client
}

type ApiResponse struct { //nolint
	Total int64       `json:"total"`
	Data  interface{} `json:"data"`
}

// Repositories a repository items list
type Repositories struct {
	List     []string `json:"repositories"`
	NextLink string   `json:"next"` // if catalog list request with pagination response will contain next page link
}

// ImageTags a tags items list
type ImageTags struct {
	Name     string   `json:"name"`
	Tags     []string `json:"tags"`
	NextLink string   // if catalog list request with pagination response will contain next page link
}

// ManifestSchemaV2 is V2 format schema for docker image manifest file which contain information about docker image, such as layers, size, and digest
// https://docs.docker.com/registry/spec/manifest-v2-2/#image-manifest-field-descriptions
type ManifestSchemaV2 struct {
	SchemaVersion     int                 `json:"schemaVersion"`
	MediaType         string              `json:"mediaType"`
	ConfigDescriptor  schema2Descriptor   `json:"config"`
	LayersDescriptors []schema2Descriptor `json:"layers"`

	// additional fields which not include in schema specification and need for this service only
	TotalSize     int64  `json:"total_size"`     // total compressed size of image data
	ContentDigest string `json:"content_digest"` // a main content digest using for delete image from registry
}

type schema2Descriptor struct {
	MediaType string   `json:"mediaType"`
	Size      int64    `json:"size"`
	Digest    string   `json:"digest"`
	URLs      []string `json:"urls,omitempty"`
}

// NewRegistry is main constructor for create registry access API instance
func NewRegistry(login, password string, settings Settings) (*Registry, error) {

	var r = new(Registry)

	r.settings = settings

	if r.settings.AuthType == Basic && login == "" {
		return nil, errors.New("login for access to registry set required when basic auth type is defined")
	}

	r.settings.credentials.login = login
	r.settings.credentials.password = password
	r.htpasswd = &htpasswd{path: settings.HtpasswdPath}

	r.httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	// it's need for self-hosted docker registry auth service with self-signed certificates
	if strings.HasPrefix(r.settings.Host, "https:") {

		transport := &http.Transport{}
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: r.settings.InsecureRequest, //nolint:gosec    // it's need  for self-signed certificate which use for https
		}
		r.httpClient.Transport = transport
	}

	// checking for at least one field of certs path is filled, other fields must require filled too
	v := reflect.ValueOf(settings.CertificatesPaths)
	var certsPathIsFilled bool
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i).Interface()

		// skips check for IP and FQDNs
		fName := v.Type().Field(i).Name
		if fName == "IP" || fName == "FQDNs" {
			continue
		}

		switch val := field.(type) {
		case string:
			if val == "" && certsPathIsFilled {
				return nil, errors.New("all fields of certificate path value required if at least on is defined")
			}

			// if filled only last field of list, but previously fields not filled
			if i == v.NumField()-1 && val != "" && !certsPathIsFilled {
				return nil, errors.New("all fields of certificate path value required if at least on is defined")
			}
			if val != "" {
				certsPathIsFilled = true
			}

		}
	}

	if r.settings.AuthType == SelfToken {

		r.htpasswd = nil // not needed for token auth
		var err error
		if certsPathIsFilled {
			if r.registryToken, err = NewRegistryToken(TokenIssuer(settings.Issuer), CertsName(settings.CertificatesPaths)); err != nil {
				return nil, err
			}
		} else {
			r.registryToken, err = NewRegistryToken(TokenIssuer(settings.Issuer))
			if err != nil {
				return nil, err
			}
		}
	}

	// try to create secure http client transport with defined certificates path
	// call this after token creation attempt, because it will create a new certificate if it doesn't exist and token auth defined
	if certsPathIsFilled {
		transport, err := createHTTPSTransport(settings.CertificatesPaths)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig.InsecureSkipVerify = r.settings.InsecureRequest

		// certificates generated for token maybe not used for TLS access and then TLS certificate should be defined
		// separately for add to trusted RootCAs
		if r.settings.HTTPSCert != "" {
			httpsCert, errRead := os.ReadFile(r.settings.HTTPSCert)
			if errRead != nil {
				return nil, fmt.Errorf("failed to add HTTPSCert to trusted root: %v", errRead)
			}
			transport.TLSClientConfig.RootCAs.AppendCertsFromPEM(httpsCert)
		}

		r.httpClient.Transport = transport
	}

	return r, nil
}

// Login implement authorization to remote registry instance with token request
func (r *Registry) Login(user store.User) (string, error) {
	authRequest := TokenRequest{
		Account: user.Login,
		Service: r.settings.Service,
	}
	return r.Token(authRequest)
}

// Token create jwt token with claims for send as response to docker registry service
// This method should call after credentials check at a high level api
func (r *Registry) Token(authRequest TokenRequest) (string, error) {

	clientToken, errToken := r.registryToken.generate(&authRequest)
	if errToken != nil {
		return "", errToken
	}

	tokenBytes, err := json.Marshal(clientToken)
	if err != nil {
		return "", err
	}
	return string(tokenBytes), nil
}

// UpdateHtpasswd update user access list every time when user add/update/delete
func (r *Registry) UpdateHtpasswd(usersFn FetchUsers) error {

	// skip update a .htpasswd file if selfToken auth is using
	if r.htpasswd == nil {
		return nil
	}

	if usersFn == nil {
		return errors.New("can't fetch users list because userFn should be defined")
	}
	users, err := usersFn.Users()
	if err != nil {
		return fmt.Errorf("can't fetch users list: %v", err)
	}

	return r.htpasswd.update(users)
}

// APIVersionCheck a minimal endpoint, mounted at /v2/ will provide version support information based on its response statuses.
// more details by link https://docs.docker.com/registry/spec/api/#api-version-check
func (r *Registry) APIVersionCheck(ctx context.Context) error {
	var apiError APIError
	baseURL := fmt.Sprintf("%s:%d/v2/", r.settings.Host, r.settings.Port)

	resp, err := r.newHTTPRequest(ctx, baseURL, "GET", nil)
	if err != nil {
		apiError.Message = fmt.Sprintf("failed to request to registry host %s", r.settings.Host)
		return err
	}

	_ = resp.Body.Close()
	if resp.StatusCode >= 400 {
		apiError.Message = fmt.Sprintf("api return error code: %d", resp.StatusCode)
		return apiError
	}

	return nil
}

// GetBlob retrieve the blob from the registry identified by digest. A HEAD request can also be issued to this endpoint
// to obtain resource information without receiving all data.
func (r *Registry) GetBlob(ctx context.Context, name, digest string) (blob []byte, err error) {
	baseURL := fmt.Sprintf("%s:%d/v2/%s/blobs/%s", r.settings.Host, r.settings.Port, name, digest)

	resp, err := r.newHTTPRequest(ctx, baseURL, "GET", nil)
	if err != nil {
		return nil, fmt.Errorf("failed get blob data err: %v", err)
	}

	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}

	blob, err = io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("api return error code: %d\n %s", resp.StatusCode, blob)
	}

	return blob, err
}

// Catalog return list a set of available repositories in the local registry cluster.
func (r *Registry) Catalog(ctx context.Context, n, last string) (Repositories, error) {
	var repos Repositories

	baseURL := fmt.Sprintf("%s:%d/v2/_catalog", r.settings.Host, r.settings.Port)

	if n != "" {
		baseURL = fmt.Sprintf("%s:%d/v2/_catalog?n=%s&last=%s", r.settings.Host, r.settings.Port, n, last)
	}

	resp, err := r.newHTTPRequest(ctx, baseURL, "GET", nil)
	if err != nil {
		return repos, err
	}
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}

	if resp.StatusCode >= 400 {
		return repos, fmt.Errorf("api return error code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&repos)
	if err != nil {
		return repos, err
	}

	nextLink, err := getPaginationNextLink(resp)
	if err != nil {
		return repos, err
	}
	repos.NextLink = nextLink

	return repos, nil
}

// ListingImageTags retrieve information about tags.
func (r *Registry) ListingImageTags(ctx context.Context, repoName, n, last string) (ImageTags, error) {
	var tags ImageTags

	baseURL := fmt.Sprintf("%s:%d/v2/%s/tags/list", r.settings.Host, r.settings.Port, repoName)

	// pagination request
	if n != "" {
		baseURL = fmt.Sprintf("%s:%d/v2/%s/tags/list?n=%s&last=%s", r.settings.Host, r.settings.Port, repoName, n, last)
	}

	resp, err := r.newHTTPRequest(ctx, baseURL, "GET", nil)
	if err != nil {
		return tags, err
	}
	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}

	if resp.StatusCode >= 400 {
		return tags, fmt.Errorf("api return error code: %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&tags)
	if err != nil {
		return tags, err
	}

	if n != "" {
		nextLink, err := getPaginationNextLink(resp)
		if err != nil {
			return tags, err
		}
		tags.NextLink = nextLink
	}

	return tags, nil
}

// Manifest do fetch the manifest identified by 'name' and 'reference' where 'reference' can be a tag or digest.
func (r *Registry) Manifest(ctx context.Context, repoName, tag string) (ManifestSchemaV2, error) {
	var manifest ManifestSchemaV2
	var apiError APIError
	baseURL := fmt.Sprintf("%s:%d/v2/%s/manifests/%s", r.settings.Host, r.settings.Port, repoName, tag)

	resp, err := r.newHTTPRequest(ctx, baseURL, "GET", nil)
	if err != nil {
		return manifest, createAPIError("failed to make request for docker registry manifest", err.Error())
	}

	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}

	if resp.StatusCode >= 400 {
		if resp != nil {
			err = json.NewDecoder(resp.Body).Decode(&apiError)
			if err != nil {
				return manifest, createAPIError("failed to parse request body with manifest fetch error", err.Error())
			}
		}
		if resp.StatusCode == http.StatusNotFound {
			return manifest, createAPIError("resource not found", "")
		}
		return manifest, apiError
	}

	err = json.NewDecoder(resp.Body).Decode(&manifest)
	if err != nil {
		return manifest, createAPIError("failed to parse request body with manifest data", err.Error())
	}

	manifest.calculateCompressedImageSize()
	manifest.ContentDigest = resp.Header.Get(contentDigestHeader)

	return manifest, nil
}

// DeleteTag will delete the manifest identified by name and reference. Note that a manifest can only be deleted by digest.
// A digest can be fetched from manifest get response header 'docker-content-digest'
func (r *Registry) DeleteTag(ctx context.Context, repoName, digest string) error {
	var apiError APIError
	baseURL := fmt.Sprintf("%s:%d/v2/%s/manifests/%s", r.settings.Host, r.settings.Port, repoName, digest)

	resp, err := r.newHTTPRequest(ctx, baseURL, "DELETE", nil)
	if err != nil {
		return createAPIError("failed to make request for delete docker registry manifest", err.Error())
	}

	if resp != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}

	if resp.StatusCode >= 400 {
		if resp != nil {
			err = json.NewDecoder(resp.Body).Decode(&apiError)
			if err != nil {
				return createAPIError("failed to parse request body when manifest delete", err.Error())
			}
		}
		if resp.StatusCode == http.StatusNotFound {
			return createAPIError("resource not found", repoName)
		}
		return apiError
	}

	return nil
}

func createHTTPSTransport(certs Certs) (*http.Transport, error) {

	certData, err := os.ReadFile(certs.CARootPath)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(certData)

	transport := &http.Transport{}
	transport.TLSClientConfig = &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    caCertPool,
	}
	return transport, nil
}

// newHTTPRequest prepare http client and execute a request to docker registry api
//
//nolint:unparam // body pass as pointer for retrieve data from response in caller method
func (r *Registry) newHTTPRequest(ctx context.Context, targetURL, method string, body []byte) (*http.Response, error) {

	req, err := http.NewRequestWithContext(ctx, method, targetURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", manifestSchemeV2)

	if r.settings.AuthType == SelfToken {
		return r.newHTTPRequestWithToken(req)
	}

	req.SetBasicAuth(r.settings.credentials.login, r.settings.credentials.password)
	return r.httpClient.Do(req)

}

// newHTTPRequestWithToken execute
func (r *Registry) newHTTPRequestWithToken(request *http.Request) (*http.Response, error) {

	resp, errReq := r.httpClient.Do(request)
	if errReq != nil {
		return nil, errReq
	}

	if resp.StatusCode != http.StatusUnauthorized {
		return resp, nil
	}

	authReq, errParse := r.ParseAuthenticateHeaderRequest(resp.Header.Get("Www-Authenticate"))
	if errParse != nil {
		return nil, errParse
	}

	tokenString, errToken := r.Token(authReq)
	if errToken != nil {
		return nil, errToken
	}

	token, err := r.registryToken.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Authorization", "Bearer "+token.Token)
	if err != nil {
		return nil, err
	}
	return r.httpClient.Do(request)

}

// getPaginationNextLink extract link for result pagination
// Compliant client implementations should always use the Link header value when proceeding through results linearly.
// The client may construct URLs to skip forward in the catalog.
//
// To get the next result set, a client would issue the request as follows, using the URL encoded in the described Link header:
//
//	GET /v2/_catalog?n=<n from the request>&last=<last repository value from previous response>
//
// The URL for the next block is encoded in RFC 5988 (https://tools.ietf.org/html/rfc5988#section-5)
func getPaginationNextLink(resp *http.Response) (string, error) {
	var nextLinkRE = regexp.MustCompile(`^ *<?([^;>]+)>? *(?:;[^;]*)*; *rel="?next"?(?:;.*)?`)

	for _, link := range resp.Header[http.CanonicalHeaderKey("Link")] {
		parts := nextLinkRE.FindStringSubmatch(link)
		if parts != nil {
			return parts[1], nil
		}
	}
	return "", ErrNoMorePages
}

// ParseAuthenticateHeaderRequest will parse 'Www-Authenticate' header for extract token authorization data.
// Header value should be like this: Bearer realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:samalba/my-app:pull,push"
// Input parameter 'access' contain data of access to resource for a user.
// Method has public access for use in tests where registry mock interface use it.
func (r *Registry) ParseAuthenticateHeaderRequest(headerValue string) (authRequest TokenRequest, err error) {
	// realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:samalba/my-app:pull,push"
	var re = regexp.MustCompile(`(\w+)=("[^"]*")`)
	var isMatched bool
	for _, match := range re.FindAllString(headerValue, -1) {
		keyValue := strings.Split(match, "=")
		if len(keyValue) != 2 {
			return authRequest, fmt.Errorf("failed to parse key/value: %v", keyValue)
		}
		key := keyValue[0]
		value := keyValue[1]
		value = strings.Trim(value, `"`)
		switch key {

		// case "realm":
		// not implemented yet.
		// on this step should match this service auth service with realm auth url

		case "service":
			authRequest.Service = value
			isMatched = true
		case "scope":
			scope := strings.Split(value, ":")
			if len(scope) != 3 {
				return authRequest, fmt.Errorf("failed to parse scope value: %s", value)
			}

			authRequest.Type = scope[0]
			authRequest.Name = scope[1]
			authRequest.Actions = strings.Split(scope[2], ",")
			isMatched = true
		}

	}
	if !isMatched {
		return authRequest, fmt.Errorf("not found header for parse token request : %s", headerValue)
	}
	return authRequest, err
}

// calculateCompressedImageSize will iterate with image layers in fetched manifest file and append size of each layers to TotalSize field
func (m *ManifestSchemaV2) calculateCompressedImageSize() {

	for _, v := range m.LayersDescriptors {
		m.TotalSize += v.Size
	}
}

// ParseURLForNextLink check pagination cursor for next
func ParseURLForNextLink(nextLink string) (next, last string, err error) {
	urlQuery, errParse := url.Parse(nextLink)
	if errParse != nil {
		return "", "", errParse
	}

	result, errParse := url.ParseQuery(urlQuery.RawQuery)
	if errParse != nil {
		return "", "", errParse
	}
	next = result.Get("n")
	last = result.Get("last")
	if next == "" && last == "" {
		return "", "", errors.New("page index is undefined in url params")
	}
	return next, last, nil
}

// APIError contain detail in their relevant sections,
// are reported as part of 4xx responses, in a json response body.
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail"`
}

// Error implement error type interface
func (ae APIError) Error() string {
	return fmt.Sprintf("%s: %s: %v", ae.Code, ae.Message, ae.Detail)
}

func createAPIError(msg, detail string) *APIError {
	return &APIError{
		Code:    "UNAVAILABLE",
		Message: msg,
		Detail:  map[string]string{"error": detail},
	}
}
