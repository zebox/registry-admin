package service

import (
	"context"
	"encoding/json"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"time"

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
		return errors.New("repository sync currently running")
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

			if err != nil && err != registry.ErrNoMorePages {
				log.Printf("failed to fetch catalog list: %v", err)
				log.Printf("[ERROR] sync operation aborted")
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
					log.Printf("[ERROR] sync operation aborted")
					return
				}

				totalTags += uint64(len(tags.Tags))

				for _, tag := range tags.Tags {
					log.Printf("[DEBUG] Tag name: %s\n", tag)
					manifest, errManifest := ds.Registry.Manifest(ctx, repo, tag)
					if errManifest != nil {
						log.Printf("[ERROR] failed to fetch manifest from repo '%s' for tag '%s' err: %s", repo, tag, errManifest)
						log.Printf("[ERROR] sync operation aborted")
						return
					}
					log.Printf("[DEBUG] Manifest: %v\n", manifest)
					rawManifestData, errMarshal := json.Marshal(&manifest)
					if errMarshal != nil {
						log.Printf("[ERROR] failed to marshal manifest data from repo '%s' for tag '%s' err: %s", repo, tag, errMarshal)
					}

					entry := &store.RegistryEntry{
						RepositoryName: repo,
						Tag:            tag,
						Digest:         manifest.ContentDigest,
						Size:           manifest.TotalSize,
						Timestamp:      time.Now().Unix(),
						Raw:            rawManifestData,
					}
					if errCreate := ds.Storage.CreateRepository(ctx, entry); err != nil {
						if errCreate == sqlite3.ErrConstraintUnique {
							log.Printf("[ERROR] failed to marshal manifest data from repo '%s' for tag '%s' err: %s", repo, tag, errCreate)
							log.Printf("[ERROR] sync operation aborted")
							return
						}
						log.Printf("[WARN] entry already exist and skipped for add: repo: '%s', tag: '%s'")
						continue
					}
					log.Printf("[DEBUG] New entry added: repo: '%s', tag: '%s'", repo, tag)
				}

				if errors.Is(errTags, registry.ErrNoMorePages) {
					log.Printf("[DEBUG] Tags for '%s' synced. Total: %d\n", repo, totalTags)
					continue
				}

				_, lastTag, err = registry.ParseUrlForNextLink(tags.NextLink)
				if err != nil {
					log.Printf("[ERROR] failed to parse next link: %v", err)
					log.Printf("[ERROR] sync operation aborted")
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