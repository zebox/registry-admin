package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestDataService_SyncExistedRepositories(t *testing.T) {

	var repositoryStore = make(map[string]store.RegistryEntry)
	var errs = &errorsEmulator{} // fake errors emitter

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := generator.Intn(55-10) + 10
	t.Logf("test data count is: %d | required repositories count: %d", n, n*n)
	testSize := n

	testDS := DataService{
		Registry: prepareRegistryMock(testSize, errs),
		Storage:  prepareStorageMock(repositoryStore, errs),
	}
	require.NotNil(t, testDS)
	err := testDS.SyncExistedRepositories(ctx)
	require.NoError(t, err)

	time.Sleep(time.Millisecond * 10)
	err = testDS.SyncExistedRepositories(ctx)
	assert.Error(t, err)

	// wait until synced
	for testDS.isSyncing {
		select {
		case <-ctx.Done():
			t.Error("context timeout before sync done")
			return
		default:
			time.Sleep(time.Millisecond * 10)
		}

	}

	assert.Equal(t, testSize*testSize, len(repositoryStore))

	// test for duplicate exclude
	err = testDS.SyncExistedRepositories(ctx)
	assert.NoError(t, err)
	for _, v := range repositoryStore {
		assert.Equal(t, testDS.lastSyncTime, v.Timestamp)
	}
	// assert.Equal(t, testSize*testSize, len(repositoryStore))

	// test with fake errors
	repositoryStore = make(map[string]store.RegistryEntry) // clear data in mock registry
	errs.createError = errorCreate
	testDS.doSyncRepositories(ctx)
	assert.Equal(t, 0, len(repositoryStore))

	repositoryStore = make(map[string]store.RegistryEntry) // clear data in mock registry
	errs.manifestError = errorManifest
	testDS.doSyncRepositories(ctx)
	assert.Equal(t, 0, len(repositoryStore))

	repositoryStore = make(map[string]store.RegistryEntry) // clear data in mock registry
	errs.listError = errorList
	testDS.doSyncRepositories(ctx)
	assert.Equal(t, 0, len(repositoryStore))

	repositoryStore = make(map[string]store.RegistryEntry) // clear data in mock registry
	errs.catalogError = errorCatalog
	testDS.doSyncRepositories(ctx)
	assert.Equal(t, 0, len(repositoryStore))

}

var (
	errorCreate   = errors.New("failed to create entry in registry")
	errorManifest = errors.New("failed to get manifest data")
	errorList     = errors.New("failed to list repository tags")
	errorCatalog  = errors.New("failed to get repository list")
)

type errorsEmulator struct {
	catalogError, listError, manifestError error // errors for registry

	createError error // errors for storage
}

func prepareRegistryMock(size int, errs *errorsEmulator) *registryInterfaceMock {
	var testRepositories = make(map[string]registry.ImageTags)
	var testManifests = make(map[string]registry.ManifestSchemaV2)

	// filling test data
	for i := 0; i < size; i++ {
		repoName := "test_repo_" + strconv.Itoa(i)
		var tags []string

		for j := 0; j < size; j++ {
			tagName := "test_tag_" + strconv.Itoa(j)
			tags = append(tags, tagName)

			// prepare test manifest
			hasher := sha256.New()
			hasher.Write([]byte(tagName))
			sha := fmt.Sprintf("%x", hasher.Sum(nil))
			manifest := registry.ManifestSchemaV2{
				TotalSize:     int64(i + j),
				ContentDigest: "sha256:" + sha,
			}
			testManifests[repoName+"_"+tagName] = manifest
		}
		testRepositories[repoName] = registry.ImageTags{Name: repoName, Tags: tags}
	}

	return &registryInterfaceMock{
		CatalogFunc: func(ctx context.Context, n string, last string) (repos registry.Repositories, err error) {

			// emit fake error
			if errs.catalogError != nil {
				return repos, errs.catalogError
			}

			// sorting testRepos
			names := make([]string, 0, len(testRepositories))
			for repoName := range testRepositories {
				names = append(names, repoName)
			}

			repos.List, repos.NextLink = paginationParse(names, n, last)
			if repos.List == nil {
				return repos, errors.New("failed to parse repository")
			}
			if repos.NextLink == "" {
				return repos, registry.ErrNoMorePages
			}
			return repos, nil
		},
		ListingImageTagsFunc: func(ctx context.Context, repoName string, n string, last string) (tags registry.ImageTags, err error) {

			// emit fake error
			if errs.listError != nil {
				return tags, errs.listError
			}

			tags.Name = repoName
			if val, ok := testRepositories[repoName]; ok {
				names := make([]string, 0, len(val.Tags))
				names = append(names, val.Tags...)

				tags.Tags, tags.NextLink = paginationParse(names, n, last)
				if tags.Tags == nil {
					return tags, errors.New("failed to parse tag list")
				}
				if tags.NextLink == "" {
					return tags, registry.ErrNoMorePages
				}
				return tags, nil
			}

			return tags, errors.New("repository not found")
		},
		ManifestFunc: func(ctx context.Context, repoName string, tag string) (manifest registry.ManifestSchemaV2, _ error) {
			// emit fake error
			if errs.manifestError != nil {
				return manifest, errs.manifestError
			}

			if val, ok := testManifests[repoName+"_"+tag]; ok {
				return val, nil
			}
			return registry.ManifestSchemaV2{}, errors.New("manifest not found")
		},
	}
}

func prepareStorageMock(repositoryStore map[string]store.RegistryEntry, errs *errorsEmulator) *engine.InterfaceMock {

	return &engine.InterfaceMock{
		CreateRepositoryFunc: func(_ context.Context, entry *store.RegistryEntry) error {

			// emit fake error
			if errs.createError != nil {
				return errs.createError
			}

			entryName := entry.RepositoryName + "_" + entry.Tag
			if _, ok := repositoryStore[entryName]; ok {
				return errors.New("UNIQUE constraint error")
			}
			repositoryStore[entryName] = *entry
			return nil
		},

		UpdateRepositoryFunc: func(ctx context.Context, conditionClause map[string]interface{}, data map[string]interface{}) error {
			// search entry first
			repoName := conditionClause[store.RegistryRepositoryNameField].(string)
			tagName := conditionClause[store.RegistryTagField].(string)
			entryName := repoName + "_" + tagName

			if entry, ok := repositoryStore[entryName]; ok {
				entry.Size = data[store.RegistrySizeNameField].(int64)
				entry.Timestamp = data[store.RegistryTimestampField].(int64)

				return nil
			}
			return errors.New("entry not found")
		},
	}
}

func paginationParse(names []string, n, last string) (repos []string, next string) {
	sort.Strings(names)
	pageSize, err := strconv.Atoi(n)
	if err != nil {
		return nil, ""
	}

	counter := 1
	for _, repoName := range names {

		if last != "" && repoName != last {
			continue
		}
		if last != "" && repoName == last {
			last = ""
			continue
		}

		repos = append(repos, repoName)
		if counter == pageSize {
			next = fmt.Sprintf(`https://example.com/v2/_catalog?n=%d&last=%s; rel="next"`, pageSize, repoName)
			return repos, next
		}
		counter++
	}
	return repos, next
}
