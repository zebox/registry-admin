// This is package implement features for interacts with instances of the docker registry,
// which is a service to manage information about docker images and enable their distribution using HTTP API V2 protocol
// detailed protocol description: https://docs.docker.com/registry/spec/api
package registry

// AuthorizationRequest is the authorization request data from registry when client auth call
type AuthorizationRequest struct {
	Account string
	Service string
	Type    string
	Name    string
	IP      string
	Actions []string
}
