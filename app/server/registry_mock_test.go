// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package server

import (
	"context"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store"
	"sync"
)

// Ensure, that registryInterfaceMock does implement registryInterface.
// If this is not the case, regenerate this file with moq.
var _ registryInterface = &registryInterfaceMock{}

// registryInterfaceMock is a mock implementation of registryInterface.
//
// 	func TestSomethingThatUsesregistryInterface(t *testing.T) {
//
// 		// make and configure a mocked registryInterface
// 		mockedregistryInterface := &registryInterfaceMock{
// 			APIVersionCheckFunc: func(ctx context.Context) error {
// 				panic("mock out the APIVersionCheck method")
// 			},
// 			CatalogFunc: func(ctx context.Context, n string, last string) (registry.Repositories, error) {
// 				panic("mock out the Catalog method")
// 			},
// 			DeleteTagFunc: func(ctx context.Context, repoName string, digest string) error {
// 				panic("mock out the DeleteTag method")
// 			},
// 			GetBlobFunc: func(ctx context.Context, name string, digest string) ([]byte, error) {
// 				panic("mock out the GetBlob method")
// 			},
// 			ListingImageTagsFunc: func(ctx context.Context, repoName string, n string, last string) (registry.ImageTags, error) {
// 				panic("mock out the ListingImageTags method")
// 			},
// 			LoginFunc: func(user store.User) (string, error) {
// 				panic("mock out the Login method")
// 			},
// 			ManifestFunc: func(ctx context.Context, repoName string, tag string) (registry.ManifestSchemaV2, error) {
// 				panic("mock out the Manifest method")
// 			},
// 			ParseAuthenticateHeaderRequestFunc: func(headerValue string) (registry.TokenRequest, error) {
// 				panic("mock out the ParseAuthenticateHeaderRequest method")
// 			},
// 			TokenFunc: func(authRequest registry.TokenRequest) (string, error) {
// 				panic("mock out the Token method")
// 			},
// 			UpdateHtpasswdFunc: func(usersFn registry.FetchUsers) error {
// 				panic("mock out the UpdateHtpasswd method")
// 			},
// 		}
//
// 		// use mockedregistryInterface in code that requires registryInterface
// 		// and then make assertions.
//
// 	}
type registryInterfaceMock struct {
	// APIVersionCheckFunc mocks the APIVersionCheck method.
	APIVersionCheckFunc func(ctx context.Context) error

	// CatalogFunc mocks the Catalog method.
	CatalogFunc func(ctx context.Context, n string, last string) (registry.Repositories, error)

	// DeleteTagFunc mocks the DeleteTag method.
	DeleteTagFunc func(ctx context.Context, repoName string, digest string) error

	// GetBlobFunc mocks the GetBlob method.
	GetBlobFunc func(ctx context.Context, name string, digest string) ([]byte, error)

	// ListingImageTagsFunc mocks the ListingImageTags method.
	ListingImageTagsFunc func(ctx context.Context, repoName string, n string, last string) (registry.ImageTags, error)

	// LoginFunc mocks the Login method.
	LoginFunc func(user store.User) (string, error)

	// ManifestFunc mocks the Manifest method.
	ManifestFunc func(ctx context.Context, repoName string, tag string) (registry.ManifestSchemaV2, error)

	// ParseAuthenticateHeaderRequestFunc mocks the ParseAuthenticateHeaderRequest method.
	ParseAuthenticateHeaderRequestFunc func(headerValue string) (registry.TokenRequest, error)

	// TokenFunc mocks the Token method.
	TokenFunc func(authRequest registry.TokenRequest) (string, error)

	// UpdateHtpasswdFunc mocks the UpdateHtpasswd method.
	UpdateHtpasswdFunc func(usersFn registry.FetchUsers) error

	// calls tracks calls to the methods.
	calls struct {
		// APIVersionCheck holds details about calls to the APIVersionCheck method.
		APIVersionCheck []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
		// Catalog holds details about calls to the Catalog method.
		Catalog []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// N is the n argument value.
			N string
			// Last is the last argument value.
			Last string
		}
		// DeleteTag holds details about calls to the DeleteTag method.
		DeleteTag []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// RepoName is the repoName argument value.
			RepoName string
			// Digest is the digest argument value.
			Digest string
		}
		// GetBlob holds details about calls to the GetBlob method.
		GetBlob []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// Name is the name argument value.
			Name string
			// Digest is the digest argument value.
			Digest string
		}
		// ListingImageTags holds details about calls to the ListingImageTags method.
		ListingImageTags []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// RepoName is the repoName argument value.
			RepoName string
			// N is the n argument value.
			N string
			// Last is the last argument value.
			Last string
		}
		// Login holds details about calls to the Login method.
		Login []struct {
			// User is the user argument value.
			User store.User
		}
		// Manifest holds details about calls to the Manifest method.
		Manifest []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// RepoName is the repoName argument value.
			RepoName string
			// Tag is the tag argument value.
			Tag string
		}
		// ParseAuthenticateHeaderRequest holds details about calls to the ParseAuthenticateHeaderRequest method.
		ParseAuthenticateHeaderRequest []struct {
			// HeaderValue is the headerValue argument value.
			HeaderValue string
		}
		// Token holds details about calls to the Token method.
		Token []struct {
			// AuthRequest is the authRequest argument value.
			AuthRequest registry.TokenRequest
		}
		// UpdateHtpasswd holds details about calls to the UpdateHtpasswd method.
		UpdateHtpasswd []struct {
			// UsersFn is the usersFn argument value.
			UsersFn registry.FetchUsers
		}
	}
	lockAPIVersionCheck                sync.RWMutex
	lockCatalog                        sync.RWMutex
	lockDeleteTag                      sync.RWMutex
	lockGetBlob                        sync.RWMutex
	lockListingImageTags               sync.RWMutex
	lockLogin                          sync.RWMutex
	lockManifest                       sync.RWMutex
	lockParseAuthenticateHeaderRequest sync.RWMutex
	lockToken                          sync.RWMutex
	lockUpdateHtpasswd                 sync.RWMutex
}

// APIVersionCheck calls APIVersionCheckFunc.
func (mock *registryInterfaceMock) APIVersionCheck(ctx context.Context) error {
	if mock.APIVersionCheckFunc == nil {
		panic("registryInterfaceMock.APIVersionCheckFunc: method is nil but registryInterface.APIVersionCheck was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockAPIVersionCheck.Lock()
	mock.calls.APIVersionCheck = append(mock.calls.APIVersionCheck, callInfo)
	mock.lockAPIVersionCheck.Unlock()
	return mock.APIVersionCheckFunc(ctx)
}

// APIVersionCheckCalls gets all the calls that were made to APIVersionCheck.
// Check the length with:
//     len(mockedregistryInterface.APIVersionCheckCalls())
func (mock *registryInterfaceMock) APIVersionCheckCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockAPIVersionCheck.RLock()
	calls = mock.calls.APIVersionCheck
	mock.lockAPIVersionCheck.RUnlock()
	return calls
}

// Catalog calls CatalogFunc.
func (mock *registryInterfaceMock) Catalog(ctx context.Context, n string, last string) (registry.Repositories, error) {
	if mock.CatalogFunc == nil {
		panic("registryInterfaceMock.CatalogFunc: method is nil but registryInterface.Catalog was just called")
	}
	callInfo := struct {
		Ctx  context.Context
		N    string
		Last string
	}{
		Ctx:  ctx,
		N:    n,
		Last: last,
	}
	mock.lockCatalog.Lock()
	mock.calls.Catalog = append(mock.calls.Catalog, callInfo)
	mock.lockCatalog.Unlock()
	return mock.CatalogFunc(ctx, n, last)
}

// CatalogCalls gets all the calls that were made to Catalog.
// Check the length with:
//     len(mockedregistryInterface.CatalogCalls())
func (mock *registryInterfaceMock) CatalogCalls() []struct {
	Ctx  context.Context
	N    string
	Last string
} {
	var calls []struct {
		Ctx  context.Context
		N    string
		Last string
	}
	mock.lockCatalog.RLock()
	calls = mock.calls.Catalog
	mock.lockCatalog.RUnlock()
	return calls
}

// DeleteTag calls DeleteTagFunc.
func (mock *registryInterfaceMock) DeleteTag(ctx context.Context, repoName string, digest string) error {
	if mock.DeleteTagFunc == nil {
		panic("registryInterfaceMock.DeleteTagFunc: method is nil but registryInterface.DeleteTag was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		RepoName string
		Digest   string
	}{
		Ctx:      ctx,
		RepoName: repoName,
		Digest:   digest,
	}
	mock.lockDeleteTag.Lock()
	mock.calls.DeleteTag = append(mock.calls.DeleteTag, callInfo)
	mock.lockDeleteTag.Unlock()
	return mock.DeleteTagFunc(ctx, repoName, digest)
}

// DeleteTagCalls gets all the calls that were made to DeleteTag.
// Check the length with:
//     len(mockedregistryInterface.DeleteTagCalls())
func (mock *registryInterfaceMock) DeleteTagCalls() []struct {
	Ctx      context.Context
	RepoName string
	Digest   string
} {
	var calls []struct {
		Ctx      context.Context
		RepoName string
		Digest   string
	}
	mock.lockDeleteTag.RLock()
	calls = mock.calls.DeleteTag
	mock.lockDeleteTag.RUnlock()
	return calls
}

// GetBlob calls GetBlobFunc.
func (mock *registryInterfaceMock) GetBlob(ctx context.Context, name string, digest string) ([]byte, error) {
	if mock.GetBlobFunc == nil {
		panic("registryInterfaceMock.GetBlobFunc: method is nil but registryInterface.GetBlob was just called")
	}
	callInfo := struct {
		Ctx    context.Context
		Name   string
		Digest string
	}{
		Ctx:    ctx,
		Name:   name,
		Digest: digest,
	}
	mock.lockGetBlob.Lock()
	mock.calls.GetBlob = append(mock.calls.GetBlob, callInfo)
	mock.lockGetBlob.Unlock()
	return mock.GetBlobFunc(ctx, name, digest)
}

// GetBlobCalls gets all the calls that were made to GetBlob.
// Check the length with:
//     len(mockedregistryInterface.GetBlobCalls())
func (mock *registryInterfaceMock) GetBlobCalls() []struct {
	Ctx    context.Context
	Name   string
	Digest string
} {
	var calls []struct {
		Ctx    context.Context
		Name   string
		Digest string
	}
	mock.lockGetBlob.RLock()
	calls = mock.calls.GetBlob
	mock.lockGetBlob.RUnlock()
	return calls
}

// ListingImageTags calls ListingImageTagsFunc.
func (mock *registryInterfaceMock) ListingImageTags(ctx context.Context, repoName string, n string, last string) (registry.ImageTags, error) {
	if mock.ListingImageTagsFunc == nil {
		panic("registryInterfaceMock.ListingImageTagsFunc: method is nil but registryInterface.ListingImageTags was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		RepoName string
		N        string
		Last     string
	}{
		Ctx:      ctx,
		RepoName: repoName,
		N:        n,
		Last:     last,
	}
	mock.lockListingImageTags.Lock()
	mock.calls.ListingImageTags = append(mock.calls.ListingImageTags, callInfo)
	mock.lockListingImageTags.Unlock()
	return mock.ListingImageTagsFunc(ctx, repoName, n, last)
}

// ListingImageTagsCalls gets all the calls that were made to ListingImageTags.
// Check the length with:
//     len(mockedregistryInterface.ListingImageTagsCalls())
func (mock *registryInterfaceMock) ListingImageTagsCalls() []struct {
	Ctx      context.Context
	RepoName string
	N        string
	Last     string
} {
	var calls []struct {
		Ctx      context.Context
		RepoName string
		N        string
		Last     string
	}
	mock.lockListingImageTags.RLock()
	calls = mock.calls.ListingImageTags
	mock.lockListingImageTags.RUnlock()
	return calls
}

// Login calls LoginFunc.
func (mock *registryInterfaceMock) Login(user store.User) (string, error) {
	if mock.LoginFunc == nil {
		panic("registryInterfaceMock.LoginFunc: method is nil but registryInterface.Login was just called")
	}
	callInfo := struct {
		User store.User
	}{
		User: user,
	}
	mock.lockLogin.Lock()
	mock.calls.Login = append(mock.calls.Login, callInfo)
	mock.lockLogin.Unlock()
	return mock.LoginFunc(user)
}

// LoginCalls gets all the calls that were made to Login.
// Check the length with:
//     len(mockedregistryInterface.LoginCalls())
func (mock *registryInterfaceMock) LoginCalls() []struct {
	User store.User
} {
	var calls []struct {
		User store.User
	}
	mock.lockLogin.RLock()
	calls = mock.calls.Login
	mock.lockLogin.RUnlock()
	return calls
}

// Manifest calls ManifestFunc.
func (mock *registryInterfaceMock) Manifest(ctx context.Context, repoName string, tag string) (registry.ManifestSchemaV2, error) {
	if mock.ManifestFunc == nil {
		panic("registryInterfaceMock.ManifestFunc: method is nil but registryInterface.Manifest was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		RepoName string
		Tag      string
	}{
		Ctx:      ctx,
		RepoName: repoName,
		Tag:      tag,
	}
	mock.lockManifest.Lock()
	mock.calls.Manifest = append(mock.calls.Manifest, callInfo)
	mock.lockManifest.Unlock()
	return mock.ManifestFunc(ctx, repoName, tag)
}

// ManifestCalls gets all the calls that were made to Manifest.
// Check the length with:
//     len(mockedregistryInterface.ManifestCalls())
func (mock *registryInterfaceMock) ManifestCalls() []struct {
	Ctx      context.Context
	RepoName string
	Tag      string
} {
	var calls []struct {
		Ctx      context.Context
		RepoName string
		Tag      string
	}
	mock.lockManifest.RLock()
	calls = mock.calls.Manifest
	mock.lockManifest.RUnlock()
	return calls
}

// ParseAuthenticateHeaderRequest calls ParseAuthenticateHeaderRequestFunc.
func (mock *registryInterfaceMock) ParseAuthenticateHeaderRequest(headerValue string) (registry.TokenRequest, error) {
	if mock.ParseAuthenticateHeaderRequestFunc == nil {
		panic("registryInterfaceMock.ParseAuthenticateHeaderRequestFunc: method is nil but registryInterface.ParseAuthenticateHeaderRequest was just called")
	}
	callInfo := struct {
		HeaderValue string
	}{
		HeaderValue: headerValue,
	}
	mock.lockParseAuthenticateHeaderRequest.Lock()
	mock.calls.ParseAuthenticateHeaderRequest = append(mock.calls.ParseAuthenticateHeaderRequest, callInfo)
	mock.lockParseAuthenticateHeaderRequest.Unlock()
	return mock.ParseAuthenticateHeaderRequestFunc(headerValue)
}

// ParseAuthenticateHeaderRequestCalls gets all the calls that were made to ParseAuthenticateHeaderRequest.
// Check the length with:
//     len(mockedregistryInterface.ParseAuthenticateHeaderRequestCalls())
func (mock *registryInterfaceMock) ParseAuthenticateHeaderRequestCalls() []struct {
	HeaderValue string
} {
	var calls []struct {
		HeaderValue string
	}
	mock.lockParseAuthenticateHeaderRequest.RLock()
	calls = mock.calls.ParseAuthenticateHeaderRequest
	mock.lockParseAuthenticateHeaderRequest.RUnlock()
	return calls
}

// Token calls TokenFunc.
func (mock *registryInterfaceMock) Token(authRequest registry.TokenRequest) (string, error) {
	if mock.TokenFunc == nil {
		panic("registryInterfaceMock.TokenFunc: method is nil but registryInterface.Token was just called")
	}
	callInfo := struct {
		AuthRequest registry.TokenRequest
	}{
		AuthRequest: authRequest,
	}
	mock.lockToken.Lock()
	mock.calls.Token = append(mock.calls.Token, callInfo)
	mock.lockToken.Unlock()
	return mock.TokenFunc(authRequest)
}

// TokenCalls gets all the calls that were made to Token.
// Check the length with:
//     len(mockedregistryInterface.TokenCalls())
func (mock *registryInterfaceMock) TokenCalls() []struct {
	AuthRequest registry.TokenRequest
} {
	var calls []struct {
		AuthRequest registry.TokenRequest
	}
	mock.lockToken.RLock()
	calls = mock.calls.Token
	mock.lockToken.RUnlock()
	return calls
}

// UpdateHtpasswd calls UpdateHtpasswdFunc.
func (mock *registryInterfaceMock) UpdateHtpasswd(usersFn registry.FetchUsers) error {
	if mock.UpdateHtpasswdFunc == nil {
		panic("registryInterfaceMock.UpdateHtpasswdFunc: method is nil but registryInterface.UpdateHtpasswd was just called")
	}
	callInfo := struct {
		UsersFn registry.FetchUsers
	}{
		UsersFn: usersFn,
	}
	mock.lockUpdateHtpasswd.Lock()
	mock.calls.UpdateHtpasswd = append(mock.calls.UpdateHtpasswd, callInfo)
	mock.lockUpdateHtpasswd.Unlock()
	return mock.UpdateHtpasswdFunc(usersFn)
}

// UpdateHtpasswdCalls gets all the calls that were made to UpdateHtpasswd.
// Check the length with:
//     len(mockedregistryInterface.UpdateHtpasswdCalls())
func (mock *registryInterfaceMock) UpdateHtpasswdCalls() []struct {
	UsersFn registry.FetchUsers
} {
	var calls []struct {
		UsersFn registry.FetchUsers
	}
	mock.lockUpdateHtpasswd.RLock()
	calls = mock.calls.UpdateHtpasswd
	mock.lockUpdateHtpasswd.RUnlock()
	return calls
}
