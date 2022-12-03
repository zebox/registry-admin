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

const (
	keyGcValue       = "ctx_gc_value"
	keyRepoGcError   = "repo_gc_error"
	keyAccessGcError = "access_gc_error"
)

type ctxKeyGC string
type ctxValueGC map[string]bool

var (
	ctxKey          ctxKeyGC = keyGcValue
	repositoryStore          = make(map[string]store.RegistryEntry)
	errs                     = &errorsEmulator{} // fake errors emitter
)

func TestDataService_SyncExistedRepositories(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	generator := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec
	n := generator.Intn(55-10) + 10
	t.Logf("test data count is: %d | required repositories count: %d", n, n*n)
	testSize := n

	testDS := &DataService{
		Registry: prepareRegistryMock(testSize),
		Storage:  prepareStorageMock(repositoryStore),
	}
	testDS.RepositoriesMaintenance(ctx, 5)

	require.NotNil(t, testDS)
	t.Log("start main syncing")
	err := testDS.SyncExistedRepositories(ctx)
	require.NoError(t, err)

	time.Sleep(time.Second * 10)

	// wait until synced
	for testDS.isWorking.Load().(bool) {
		errSync := testDS.SyncExistedRepositories(ctx)
		assert.Error(t, errSync)
		time.Sleep(time.Millisecond * 100)
	}

	assert.Equal(t, testSize*testSize, len(repositoryStore))

	t.Log("test for duplicate exclude")
	err = testDS.SyncExistedRepositories(ctx)
	assert.NoError(t, err)

	lastSync := testDS.lastSyncDate.Load().(int64)
	for _, v := range repositoryStore {
		assert.Equal(t, lastSync, v.Timestamp)
	}
	ctx.Done()

	t.Log("test with fake errors")
	repositoryStore = make(map[string]store.RegistryEntry) // clear data in mock registry

	ctx = context.WithValue(context.Background(), ctxKey, errorCreate)
	testDS.doSyncRepositories(ctx)
	assert.Equal(t, 0, len(repositoryStore))
	assert.Equal(t, errorCreate, errs.currentError)

	repositoryStore = make(map[string]store.RegistryEntry) // clear data in mock registry
	ctx = context.WithValue(context.Background(), ctxKey, errorManifest)
	testDS.doSyncRepositories(ctx)
	assert.Equal(t, 0, len(repositoryStore))
	assert.Equal(t, errorManifest, errs.currentError)

	repositoryStore = make(map[string]store.RegistryEntry) // clear data in mock registry
	ctx = context.WithValue(context.Background(), ctxKey, errorList)
	testDS.doSyncRepositories(ctx)
	assert.Equal(t, 0, len(repositoryStore))
	assert.Equal(t, errorList, errs.currentError)

	repositoryStore = make(map[string]store.RegistryEntry) // clear data in mock registry
	ctx = context.WithValue(context.Background(), ctxKey, errorCatalog)
	testDS.doSyncRepositories(ctx)
	assert.Equal(t, 0, len(repositoryStore))
	assert.Equal(t, errorCatalog, errs.currentError)
}

func TestDataService_RepositoriesMaintaining(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*6)
	defer cancel()

	generator := rand.New(rand.NewSource(time.Now().UnixNano())) // #nosec
	n := generator.Intn(20-10) + 10
	t.Logf("test data count is: %d | required repositories count: %d", n, n*n)
	testSize := n

	testDS := &DataService{
		Registry: prepareRegistryMock(testSize),
		Storage:  prepareStorageMock(repositoryStore),
	}

	ctx = context.WithValue(ctx, ctxKey, ctxValueGC{keyRepoGcError: false, keyAccessGcError: false})
	testDS.RepositoriesMaintenance(ctx, 2)
	<-ctx.Done()

	ctxValue := ctx.Value(ctxKey)
	require.NotNil(t, ctxValue)
	require.IsType(t, ctxValueGC{}, ctxValue)

	ctxValue.(ctxValueGC)[keyRepoGcError] = false
	err := testDS.doGarbageCollector(ctx)
	assert.NotNil(t, err)

	testDS.lastSyncDate.Store(int64(0))
	err = testDS.doGarbageCollector(ctx)
	assert.Equal(t, ErrNoSyncedYet, err)
}

var (
	errorCreate   = errors.New("failed to create entry in registry")
	errorManifest = errors.New("failed to get manifest data")
	errorList     = errors.New("failed to list repository tags")
	errorCatalog  = errors.New("failed to get repository list")
)

type errorsEmulator struct {
	currentError error // errors for registry
}

func ctxCheckErrorFn(ctx context.Context, err error) error {
	if ctxValue := ctx.Value(ctxKey); ctxValue != nil {
		if ctxValue.(error) == err {
			errs.currentError = err
			return err
		}
	}
	return nil
}

func prepareRegistryMock(size int) *registryInterfaceMock {
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
			if err := ctxCheckErrorFn(ctx, errorCatalog); err != nil {
				return repos, errorCatalog
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
		ListingImageTagsFunc: func(ctx context.Context, repoName, n, last string) (tags registry.ImageTags, err error) {

			// emit fake error
			if err := ctxCheckErrorFn(ctx, errorList); err != nil {
				return tags, errorList
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
			if err := ctxCheckErrorFn(ctx, errorManifest); err != nil {
				return manifest, errorManifest
			}

			if val, ok := testManifests[repoName+"_"+tag]; ok {
				return val, nil
			}
			return registry.ManifestSchemaV2{}, errors.New("manifest not found")
		},
	}
}

func prepareStorageMock(repositoryStore map[string]store.RegistryEntry) *engine.InterfaceMock {

	ctxCheckFn := func(ctx context.Context, key string) error {
		ctxValue := ctx.Value(ctxKey)
		if ctxValue != nil {
			switch val := ctxValue.(type) { // nolint
			case ctxValueGC:
				if ok := val[keyRepoGcError]; !ok {
					val[keyRepoGcError] = true // invert value for emit error at next call
					return nil
				}
			}
		}
		return errors.New("no task for garbage collector")
	}

	return &engine.InterfaceMock{
		CreateRepositoryFunc: func(ctx context.Context, entry *store.RegistryEntry) error {

			// emit fake error
			if err := ctxCheckErrorFn(ctx, errorCreate); err != nil {
				return err
			}

			entryName := entry.RepositoryName + "_" + entry.Tag
			if _, ok := repositoryStore[entryName]; ok {
				return errors.New("UNIQUE constraint error")
			}
			repositoryStore[entryName] = *entry
			return nil
		},

		UpdateRepositoryFunc: func(ctx context.Context, conditionClause map[string]interface{}, data map[string]interface{}) error { //nolint:lll

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

		RepositoryGarbageCollectorFunc: func(ctx context.Context, syncDate int64) error {
			return ctxCheckFn(ctx, keyRepoGcError)
		},

		AccessGarbageCollectorFunc: func(ctx context.Context) error {
			return ctxCheckFn(ctx, keyAccessGcError)
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
