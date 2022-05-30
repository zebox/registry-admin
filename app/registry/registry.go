package registry

import (
	"errors"
	"reflect"
)

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
	SelfToken                 // define this service as main auth/authz server for docker registry host
)

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
}

// Registry is main instance for manipulation access of self-hosted docker registry
type Registry struct {
	settings      Settings
	registryToken *registryToken
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
