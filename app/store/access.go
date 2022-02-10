// Access is the main instance of object which define access right for Users and Groups

package store

// AccessInterface defines methods provided for Access manipulation
type AccessInterface interface {
	// CheckAccessPermit will check access for a specify entry
	CheckAccessPermit() (bool, error)
}

// Access model for accessing rules
// Fields of access object defined with docker docs - https://docs.docker.com/registry/spec/auth/scope
type Access struct {
	ID      int64  `json:"id"`
	Owner   int64  `json:"owner_id"`
	IsGroup bool   `json:"is_group"`
	Name    string `json:"name"`

	// Represents type of resource which the resource name is intended.
	// Available next resource types:
	// 		- repository - represents a single repository within a registry;
	//		- registry - represents the entire registry (Used for administrative actions or lookup operations that span an entire registry).
	Type string `json:"type"`

	// The resource name represent the name which identifies a resource for a resource provider
	ResourceName string `json:"resource_name"`

	// The resource actions define the actions which the access token allows to be performed on the identified resource.
	// These actions are type specific but will normally have actions identifying read and write access on the resource.
	// Example for the 'repository' type are 'pull' for read access and 'push' for write access.
	Action string `json:"action"`

	// Marks a access item as disabled
	Disabled bool `json:"disabled"`
}
