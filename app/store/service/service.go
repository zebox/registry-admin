package service

import (
	"context"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store/engine"

	log "github.com/go-pkgz/lgr"
)

const defaultPageSize = "50" // the number of repository items for pagination request when catalog listing

// registryInterface implement method for access data of a registry instance
type registryInterface interface {
	// Catalog return list a set of available repositories in the local registry cluster.
	Catalog(ctx context.Context, n, last string) (registry.Repositories, error)

	// ListingImageTags retrieve information about tags.
	ListingImageTags(ctx context.Context, repoName, n, last string) (registry.ImageTags, error)

	// Manifest will fetch the manifest identified by 'name' and 'reference' where 'reference' can be a tag or digest.
	Manifest(ctx context.Context, repoName, tag string) (registry.ManifestSchemaV2, error)
}

// DataService is service which allow manipulation entries of registry such repositories or tags
type DataService struct {
	Registry registryInterface
	Storage  engine.Interface

	isSyncing bool // used for checks status syncing operation is active
}

// SyncExistedRepositories will check existed entries at a registry service and synchronize it
func (ds *DataService) SyncExistedRepositories(ctx context.Context) error {

	// prevent parallel syncing
	if ds.isSyncing {
		return errors.New("repository syncing currently already running")
	}
	ds.isSyncing = true
	defer func() { ds.isSyncing = false }()

	var (
		n          = defaultPageSize // item number per page
		lastRepo   string
		totalRepos uint64
	)

	go func() {
		for {
			repos, err := ds.Registry.Catalog(context.Background(), n, lastRepo)

			if err != nil && !errors.Is(err, registry.ErrNoMorePages) {
				log.Printf("failed to syncing existed repositories: %v", err)
				break
			}

			totalRepos += uint64(len(repos.List))

			for _, repo := range repos.List {
				var (
					lastTag   string
					totalTags uint64
				)
				log.Printf("[DEBUG] Repository name: %s\n", repo)

				tags, errTags := ds.Registry.ListingImageTags(ctx, repo, defaultPageSize, lastTag)
				if errTags != nil && !errors.Is(errTags, registry.ErrNoMorePages) {
					log.Printf("[ERROR] failed to repository tags at repository '%s': %v", repo, errTags)
					return
				}

				totalTags += uint64(len(tags.Tags))

				for _, tag := range tags.Tags {
					log.Printf("[DEBUG] Tag name: %s\n", tag)
					manifest, errManifest := ds.Registry.Manifest(ctx, repo, tag)
					if errManifest != nil {
						return
					}
					log.Printf("[DEBUG] Manifest: %v\n", manifest)
				}

				if errors.Is(errTags, registry.ErrNoMorePages) {
					log.Printf("[DEBUG] Tags for '%s' synced. Total: %d\n", repo, totalTags)
					continue
				}

				_, lastTag, err = registry.ParseUrlForNextLink(tags.NextLink)
				if err != nil {
					log.Printf("[ERROR] failed to parse next link: %v", err)
					break
				}
			}

			if errors.Is(err, registry.ErrNoMorePages) {
				log.Printf("[INFO] Repositories synced. Total: %d\n", totalRepos)
				break
			}

			n, lastRepo, err = registry.ParseUrlForNextLink(repos.NextLink)
			if err != nil {
				log.Printf("[ERROR] failed to parse next link: %v", err)
				break
			}
		}
	}()

	return nil
}
