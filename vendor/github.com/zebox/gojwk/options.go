package gojwk

// Main Options for JWKS
type Options func(k *Keys)

// Storage define external storage for Keys save and load.
// Save method need for save when new Keys generated Keys
func Storage(s keyStorage) Options {
	return func(k *Keys) {
		k.storage = s
	}
}

// BitSize value
func BitSize(bitSize int) Options {
	return func(k *Keys) {
		k.bitSize = bitSize
	}
}
