package registry

// This is package implement features for interacts with instances of the docker registry,
// which is a service to manage information about docker images and enable their distribution using HTTP API V2 protocol
// detailed protocol description: https://docs.docker.com/registry/spec/api

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

// authType define auth mechanism for accessing to docker registry using a docker HTTP API protocol
type authType int8

const (
	Basic     authType = iota // allow access using auth basic credentials
	Token                     // defined token for access to docker registry
	SelfToken                 // define this service as main auth/authz server for docker registry host
)

type registryTokenInterface interface {
	// Generate This method will generate JWT for auth check at self-hosted docker registry host
	Generate(authRequest *AuthorizationRequest) (clientToken, error)

	// CreateCerts will create new bundle with key, cert and CA bundle where 'self-token' registry auth type is used
	// the path input parameter define where created certs will be stored
	CreateCerts(path string) error
}

type Options struct {

	// Host is a fqdn of docker registry host
	Host string

	// define authenticate type for access to docker registry api
	AuthType authType

	// define path to for keys bundles
	Key    string // is a private key
	Cert   string // is a public key
	CARoot string // is CA root bundle
}

// Registry is main instance for manipulation access of self-hosted docker registry
type Registry struct {
	Options
	registryToken registryTokenInterface
}
