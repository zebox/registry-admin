// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package service

import (
	"context"
	"github.com/zebox/registry-admin/app/registry"
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
// 			CatalogFunc: func(ctx context.Context, n string, last string) (registry.Repositories, error) {
// 				panic("mock out the Catalog method")
// 			},
// 			ListingImageTagsFunc: func(ctx context.Context, repoName string, n string, last string) (registry.ImageTags, error) {
// 				panic("mock out the ListingImageTags method")
// 			},
// 			ManifestFunc: func(ctx context.Context, repoName string, tag string) (registry.ManifestSchemaV2, error) {
// 				panic("mock out the Manifest method")
// 			},
// 		}
//
// 		// use mockedregistryInterface in code that requires registryInterface
// 		// and then make assertions.
//
// 	}
type registryInterfaceMock struct {
	// CatalogFunc mocks the Catalog method.
	CatalogFunc func(ctx context.Context, n string, last string) (registry.Repositories, error)

	// ListingImageTagsFunc mocks the ListingImageTags method.
	ListingImageTagsFunc func(ctx context.Context, repoName string, n string, last string) (registry.ImageTags, error)

	// ManifestFunc mocks the Manifest method.
	ManifestFunc func(ctx context.Context, repoName string, tag string) (registry.ManifestSchemaV2, error)

	// calls tracks calls to the methods.
	calls struct {
		// Catalog holds details about calls to the Catalog method.
		Catalog []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// N is the n argument value.
			N string
			// Last is the last argument value.
			Last string
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
		// Manifest holds details about calls to the Manifest method.
		Manifest []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
			// RepoName is the repoName argument value.
			RepoName string
			// Tag is the tag argument value.
			Tag string
		}
	}
	lockCatalog          sync.RWMutex
	lockListingImageTags sync.RWMutex
	lockManifest         sync.RWMutex
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
