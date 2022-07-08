package store

import "encoding/json"

// registry implements registry entries types for store and fetch them with storage query mechanism.
// Registry V2 API listing repository list with pagination limit only, but not allow query repository item by name.
// When user need find item by name a user need listing all entries until find appropriate item.
// For fix this problem registry will notify this service about repository data change and change will be sync with storage.

// RegistryEntry is main entry  records which will save in storage
type RegistryEntry struct {
	ID             int64           `json:"id"`
	RepositoryName string          `json:"repository_name"` // Repository identifies the named repository.
	Tag            string          `json:"tag"`             // Tag provides the tag
	Digest         string          `json:"digest"`          // Digest uniquely identifies the content. A byte stream can be verified against this digest.
	Size           int64           `json:"size"`            // Size in bytes of content.
	Raw            json.RawMessage `json:"raw,omitempty"`   // Raw is a whole notify event data in json
}
