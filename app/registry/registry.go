package registry

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

const (
	Basic     authType = iota // allow access using auth basic credentials
	SelfToken                 // define this service as main auth/authz server for docker registry host
)

var (
	ErrNoMorePages = errors.New("no more pages")
)

// AuthorizationRequest is the authorization request data from registry when client auth call
// for detailed description go to https://docs.docker.com/registry/spec/auth/jwt/
type AuthorizationRequest struct {

	// Bind to 'sub' token header
	// The subject of the token; the name or id of the client which requested it.
	// This should be empty (`""`) if the client did not authenticate.
	Account string

	// Bind to token 'aud' header. The intended audience of the token; the name or id of the service which will verify
	// the token to authorize the client/subject.
	Service string

	// The subject of the token; the name or id of the client which requested it.
	// This should be empty (`""`) if the client did not authenticate.
	Type string

	// The name of the resource of the given type hosted by the service.
	Name string

	// An array of strings which give the actions authorized on this resource.
	Actions []string

	IP string
}

type Settings struct {

	// Host is a fqdn of docker registry host
	Host string

	// Port which registry accept requests
	Port int

	// define authenticate type for access to docker registry api
	AuthType authType

	// credentials define user and login pair for auth in docker registry, when auth type set as basic
	credentials struct {
		login, password string
	}

	// CertificatesPaths define a path to private, public keys and CA certificate.
	// If CertificatesPaths has all fields are empty, registryToken will create keys by default, with default path.
	// If CertificatesPaths has all fields are empty, but certificates files exist registryToken try to load existed keys and CA file.
	CertificatesPaths Certs

	// InsecureRequest define option secure for make a https request to docker registry host, false by default
	InsecureRequest bool
}

// Registry is main instance for manipulation access of self-hosted docker registry
type Registry struct {
	settings      Settings
	registryToken *registryToken
}

// ApiError contain detail in their relevant sections,
// are reported as part of 4xx responses, in a json response body.
type ApiError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Detail  interface{} `json:"detail"`
}

type ApiResponse struct {
	Total int64       `json:"total"`
	Data  interface{} `json:"data"`
}

// Repositories a repository items list
type Repositories struct {
	List     []string `json:"repositories"`
	NextLink string   // if catalog list request with pagination response will contain next page link
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
	ConfigDescriptor  Schema2Descriptor   `json:"config"`
	LayersDescriptors []Schema2Descriptor `json:"layers"`

	// additional fields which not include in schema specification and need for this service only
	TotalSize     int64  `json:"total_size"`     // total compressed size of image data
	ContentDigest string `json:"content_digest"` // a main content digest using for delete image from registry
}

type Schema2Descriptor struct {
	MediaType string   `json:"mediaType"`
	Size      int64    `json:"size"`
	Digest    string   `json:"digest"`
	URLs      []string `json:"urls,omitempty"`
}

// NewRegistry is main constructor for create registry access API instance
func NewRegistry(login, password, secret string, settings Settings) (*Registry, error) {

	var r = new(Registry)

	r.settings = settings

	if r.settings.AuthType == Basic && login == "" {
		return nil, errors.New("at least login should set when basic auth type is set")
	}

	r.settings.credentials.login = login
	r.settings.credentials.password = password

	if r.settings.AuthType == SelfToken {
		if len(secret) == 0 {
			return nil, errors.New("token secret must be defined for 'self_token' auth type")
		}

		// checking for at least one field of certs path is filled, other fields must require filled too
		v := reflect.ValueOf(settings.CertificatesPaths)
		var certsPathIsFilled bool
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i).Interface()
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

		var err error
		if certsPathIsFilled {
			r.registryToken, err = NewRegistryToken(secret, TokenIssuer(settings.Host), CertsName(settings.CertificatesPaths))
			if err != nil {
				return nil, err
			}
		} else {
			r.registryToken, err = NewRegistryToken(secret, TokenIssuer(settings.Host))
			if err != nil {
				return nil, err
			}
		}
	}

	return r, nil
}

// ApiVersionCheck a minimal endpoint, mounted at /v2/ will provide version support information based on its response statuses.
// more details by link https://docs.docker.com/registry/spec/api/#api-version-check
func (r *Registry) ApiVersionCheck(ctx context.Context) (ApiError, error) {
	var apiError ApiError
	url := fmt.Sprintf("%s:%d/v2/", r.settings.Host, r.settings.Port)
	resp, err := r.newHttpRequest(ctx, url, "GET", nil)
	if err != nil {
		apiError.Message = fmt.Sprintf("failed to request to registry host %s", r.settings.Host)
		return apiError, err
	}
	_ = resp.Body.Close()
	if resp.StatusCode >= 400 {
		apiError.Message = fmt.Sprintf("api return error code: %d", resp.StatusCode)
	}
	return apiError, nil
}

func (r *Registry) Catalog(ctx context.Context, n, last string) (Repositories, error) {
	var repos Repositories

	baseUrl := fmt.Sprintf("%s:%d/v2/_catalog", r.settings.Host, r.settings.Port)

	if n != "" {
		baseUrl = fmt.Sprintf("%s:%d/v2/_catalog?n=%s&last=%s", r.settings.Host, r.settings.Port, n, last)
	}

	resp, err := r.newHttpRequest(ctx, baseUrl, "GET", nil)
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

	// pagination request
	if n != "" {
		nextLink, err := getPaginationNextLink(resp)
		if err != nil {
			return repos, err
		}
		repos.NextLink = nextLink
	}

	return repos, nil
}

func (r *Registry) ListingImageTags(ctx context.Context, repoName, n, last string) (ImageTags, error) {
	var tags ImageTags

	baseUrl := fmt.Sprintf("%s:%d/v2/%s/tags/list", r.settings.Host, r.settings.Port, repoName)

	// pagination request
	if n != "" {
		baseUrl = fmt.Sprintf("%s:%d/v2/%s/tags/list?n=%s&last=%s", r.settings.Host, r.settings.Port, repoName, n, last)
	}

	resp, err := r.newHttpRequest(ctx, baseUrl, "GET", nil)
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

func (r *Registry) Manifest(ctx context.Context, repoName, tag string) (ManifestSchemaV2, ApiError) {
	var manifest ManifestSchemaV2
	var apiError ApiError
	baseUrl := fmt.Sprintf("%s:%d/v2/%s/manifests/%s", r.settings.Host, r.settings.Port, repoName, tag)

	resp, err := r.newHttpRequest(ctx, baseUrl, "GET", nil)
	if err != nil {
		return manifest, makeApiError("failed to make request for docker registry manifest", err.Error())
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
				return manifest, makeApiError("failed to parse request body with manifest fetch error", err.Error())
			}
		}
		if resp.StatusCode == http.StatusNotFound {
			return manifest, makeApiError("resource not found", "")
		}
		return manifest, apiError
	}

	err = json.NewDecoder(resp.Body).Decode(&manifest)
	if err != nil {
		return manifest, makeApiError("failed to parse request body with manifest data", err.Error())
	}

	manifest.calculateCompressedImageSize()
	manifest.ContentDigest = resp.Header.Get(contentDigestHeader)

	return manifest, apiError
}

// newHttpRequest prepare http client and execute a request to docker registry api
func (r Registry) newHttpRequest(ctx context.Context, url, method string, body []byte) (*http.Response, error) {

	transport := &http.Transport{}

	// it's need for self-hosted docker registry with self-signed certificates
	if strings.HasPrefix(url, "https:") {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: r.settings.InsecureRequest}
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", manifestSchemeV2)
	req.SetBasicAuth(r.settings.credentials.login, r.settings.credentials.password)

	return client.Do(req)

}

// getPaginationNextLink extract link for result pagination
// Compliant client implementations should always use the Link header value when proceeding through results linearly.
// The client may construct URLs to skip forward in the catalog.
//
// To get the next result set, a client would issue the request as follows, using the URL encoded in the described Link header:
//   	GET /v2/_catalog?n=<n from the request>&last=<last repository value from previous response>
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

// calculateCompressedImageSize will iterate with image layers in fetched manifest file and append size of each layers to TotalSize field
func (m *ManifestSchemaV2) calculateCompressedImageSize() {

	for _, v := range m.LayersDescriptors {
		m.TotalSize += v.Size
	}
}

func makeApiError(msg, detail string) ApiError {
	return ApiError{
		Code:    "-1",
		Message: msg,
		Detail:  map[string]string{"error": detail},
	}
}
