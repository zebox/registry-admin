package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/registry"
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

	for {
		repos, err := testDS.Registry.Catalog(ctx, n, last)

		n, last, err = registry.ParseUrlForNextLink(repos.NextLink)

		if errors.Is(err, registry.ErrNoMorePages) {
			t.Logf("[INFO] Repositories synced. Total: %d\n", len(repos.List))
			break
		}

		if err != nil {
			t.Logf("failed to parse next link: %v", err)
			break
		}
	}
}

func prepareRegistryMock() *registryInterfaceMock {
	var testRepositories = make(map[string]registry.ImageTags)
	var testManifests = make(map[string]registry.ManifestSchemaV2)

	// filling test data
	for i := 0; i < 60; i++ {
		repoName := "test_repo_" + strconv.Itoa(i)
		var tags []string

		for j := 0; j < 60; j++ {
			tagName := "test_tag_" + strconv.Itoa(i)
			tags = append(tags, tagName)

			// prepare test manifest
			hasher := sha256.New()
			hasher.Write([]byte(tagName))
			sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
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

			pageSize, err := strconv.Atoi(n)
			if err != nil {
				return repos, err
			}
			counter := 1

			for repoName := range testRepositories {
				if last != "" && repoName != last {
					continue
				}
				repos.List = append(repos.List, repoName)
				if counter == pageSize {
					repos.NextLink = fmt.Sprintf(`http://example.com/v2/_catalog?n=%s&last=%s; rel="next"`, n, repoName)
					return repos, nil
				}
				counter++
			}
			return repos, registry.ErrNoMorePages
		},
		ListingImageTagsFunc: func(ctx context.Context, repoName string, n string, last string) (registry.ImageTags, error) {
			return registry.ImageTags{}, nil
		},
		ManifestFunc: func(ctx context.Context, repoName string, tag string) (registry.ManifestSchemaV2, error) {
			return registry.ManifestSchemaV2{}, nil
		},
	}
}
