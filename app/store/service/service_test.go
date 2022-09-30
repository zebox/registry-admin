package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"sort"
	"strconv"
	"testing"
	"time"
)

func TestDataService_SyncExistedRepositories(t *testing.T) {

	var repositoryStore = make(map[string]store.RegistryEntry)
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	testSize := 55

	testDS := DataService{
		Registry: prepareRegistryMock(testSize),
		Storage:  prepareStorageMock(repositoryStore),
	}
	require.NotNil(t, testDS)
	//require.NoError(t, testDS.SyncExistedRepositories(ctx))
	testDS.doSyncRepositories(ctx)
	//<-ctx.Done()
	assert.Equal(t, testSize*testSize, len(repositoryStore))
	/*
		var n, last string = "20", ""
		var list registry.ImageTags
		for {
			repos, err := testDS.Registry.ListingImageTags(ctx, "test_repo_10", n, last)

			list.Tags = append(list.Tags, repos.Tags...)
			if errors.Is(err, registry.ErrNoMorePages) {
				t.Logf("[INFO] Repositories synced. Total: %d\n", len(repos.Tags))
				break
			}
			n, last, err = registry.ParseUrlForNextLink(repos.NextLink)
			if err != nil {
				t.Logf("failed to parse next link: %v", err)
				break
			}
		}
		for _, l := range list.Tags {
			m, err := testDS.Registry.Manifest(ctx, "test_repo_10", l)
			assert.NoError(t, err)
			t.Logf("%v", m)
		}*/
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
			tags.Name = repoName
			if val, ok := testRepositories[repoName]; ok {
				names := make([]string, 0, len(val.Tags))
				for _, tagName := range val.Tags {
					names = append(names, tagName)
				}

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
		ManifestFunc: func(ctx context.Context, repoName string, tag string) (registry.ManifestSchemaV2, error) {
			if val, ok := testManifests[repoName+"_"+tag]; ok {
				return val, nil
			}
			return registry.ManifestSchemaV2{}, errors.New("manifest not found")
		},
	}
}

func prepareStorageMock(repositoryStore map[string]store.RegistryEntry) *engine.InterfaceMock {

	return &engine.InterfaceMock{
		CreateRepositoryFunc: func(_ context.Context, entry *store.RegistryEntry) error {
			entryName := entry.RepositoryName + "_" + entry.Tag
			if _, ok := repositoryStore[entryName]; ok {
				return sqlite3.ErrConstraintUnique
			}
			repositoryStore[entryName] = *entry
			return nil
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
