package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/registry"
	"sort"
	"strconv"
	"testing"
)

func TestDataService_SyncExistedRepositories(t *testing.T) {
	ctx := context.Background()
	testDS := DataService{
		Registry: prepareRegistryMock(),
	}
	require.NotNil(t, testDS)

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
	}
}

func prepareRegistryMock() *registryInterfaceMock {
	var testRepositories = make(map[string]registry.ImageTags)
	var testManifests = make(map[string]registry.ManifestSchemaV2)

	// filling test data
	for i := 0; i < 55; i++ {
		repoName := "test_repo_" + strconv.Itoa(i)
		var tags []string

		for j := 0; j < 55; j++ {
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
			sort.Strings(names)

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
				sort.Strings(names)
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

func paginationParse(names []string, n, last string) (repos []string, next string) {
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
