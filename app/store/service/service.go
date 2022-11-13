package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"strings"
	"sync"
	"time"

	log "github.com/go-pkgz/lgr"
)

const defaultPageSize = "50" // the number of repository items for pagination request when catalog listing

var ErrNoSyncedYet = errors.New("garbage collector skip because sync required start first")

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

	lastSyncDate int64 // timestamp for mark actual record and use it for repository garbage collector

	// used for checks status either syncing operation or garbage collector is active
	// it prevents stacking task in queue and run in parallels
	isWorking bool

	mutex sync.RWMutex
}

// SyncExistedRepositories will check existed entries at a registry service and synchronize it
func (ds *DataService) SyncExistedRepositories(ctx context.Context) error {

	// prevent parallel syncing
	ds.mutex.RLock()
	defer ds.mutex.RUnlock()
	if ds.isWorking {
		return errors.New("repository sync currently running")
	}

	go ds.doSyncRepositories(ctx)
	return nil
}

func (ds *DataService) doSyncRepositories(ctx context.Context) {

	ds.mutex.Lock()
	ds.isWorking = true
	defer func() {
		ds.isWorking = false
		ds.mutex.Unlock()
	}()

	now := time.Now().Unix()

	var (
		n          = defaultPageSize // item number per page
		lastRepo   string
		totalRepos uint64
		lastTag    string
		totalTags  uint64
	)
	for {
		repos, errCatalog := ds.Registry.Catalog(context.Background(), n, lastRepo)

		if errCatalog != nil && errCatalog != registry.ErrNoMorePages {
			log.Printf("failed to fetch catalog list: %v", errCatalog)
			log.Printf("[ERROR] sync operation aborted")
			break
		}

		totalRepos += uint64(len(repos.List))

		for _, repo := range repos.List {

			log.Printf("[DEBUG] Repository name: %s\n", repo)

			for {
				tags, errTags := ds.Registry.ListingImageTags(ctx, repo, defaultPageSize, lastTag)
				if errTags != nil && !errors.Is(errTags, registry.ErrNoMorePages) {
					log.Printf("[ERROR] failed to get repository tags at repository '%s': %v", repo, errTags)
					log.Printf("[ERROR] sync operation aborted")
					return
				}

				totalTags += uint64(len(tags.Tags))

				if tags.Tags != nil {
					for _, tag := range tags.Tags {
						log.Printf("[DEBUG] Tag name: %s\n", tag)
						manifest, errManifest := ds.Registry.Manifest(ctx, repo, tag)
						if errManifest != nil {
							log.Printf("[ERROR] failed to fetch manifest from repo '%s' for tag '%s' errCatalog: %s", repo, tag, errManifest)
							log.Printf("[ERROR] sync operation aborted")
							return
						}
						log.Printf("[DEBUG] Manifest: %v\n", manifest)
						rawManifestData, errMarshal := json.Marshal(&manifest)
						if errMarshal != nil {
							log.Printf("[ERROR] failed to marshal manifest data from repo '%s' for tag '%s' errCatalog: %s", repo, tag, errMarshal)
						}

						entry := &store.RegistryEntry{
							RepositoryName: repo,
							Tag:            tag,
							Digest:         manifest.ConfigDescriptor.Digest,
							Size:           manifest.TotalSize,
							Timestamp:      now,
							Raw:            string(rawManifestData),
						}
						if errCreate := ds.Storage.CreateRepository(ctx, entry); errCreate != nil {
							if !strings.HasPrefix(errCreate.Error(), "UNIQUE") {
								log.Printf("[ERROR] failed to marshal manifest data from repo '%s' for tag '%s' err: %s", repo, tag, errCreate)
								log.Printf("[ERROR] sync operation aborted")
								return
							}

							// if repositories with specific tag already exist the service try update dynamic field and set a new timestamp.
							// Then timestamp using for garbage collector for detect outdated data and remove one from repository store
							log.Printf("[DEBUG] entry already exist and will update : repo: '%s', tag: '%s'", repo, tag)

							condition := map[string]interface{}{
								store.RegistryRepositoryNameField: repo,
								store.RegistryTagField:            tag,
							}

							fieldForUpdate := map[string]interface{}{
								store.RegistrySizeNameField:  manifest.TotalSize,
								store.RegistryTimestampField: now,
							}

							if errUpdate := ds.Storage.UpdateRepository(ctx, condition, fieldForUpdate); errUpdate != nil {
								log.Printf("[ERROR] sync operation aborted: repo '%s' for tag '%s' err: %s", repo, tag, errUpdate)
								return
							}
							continue
						}
						log.Printf("[DEBUG] New entry added: repo: '%s', tag: '%s'", repo, tag)
					}
				}

				if errors.Is(errTags, registry.ErrNoMorePages) {
					log.Printf("[DEBUG] Tags for '%s' synced. Total: %d\n", repo, totalTags)
					lastTag = ""
					break
				}

				_, lastTag, errTags = registry.ParseUrlForNextLink(tags.NextLink)
				if errTags != nil {
					log.Printf("[ERROR] failed to parse next link: %v", errTags)
					log.Printf("[ERROR] sync operation aborted")
					break
				}
			}
		}

		if errors.Is(errCatalog, registry.ErrNoMorePages) {
			log.Printf("[INFO] Repositories synced. Total: %d\n", totalRepos)
			break
		}

		n, lastRepo, errCatalog = registry.ParseUrlForNextLink(repos.NextLink)
		if errCatalog != nil {
			log.Printf("[ERROR] failed to parse next link: %v", errCatalog)
			break
		}
	}

	ds.lastSyncDate = now
}

// RepositoriesMaintenance check repositories for outdated or updated data in repository storage
// with 'lastSyncDate' value. Timestamp field update at every sync call in repository storage
// and compare with 'lastSyncDate' variable.
// If values above is different garbage collector will remove all outdated entries
func (ds *DataService) RepositoriesMaintenance(ctx context.Context, timeout int64) {

	if timeout == 0 {
		timeout = 60
	}
	ticker := time.NewTicker(time.Duration(timeout) * time.Hour)

	// starting garbage collector background task
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("[DEBUG] repositories maintaining task stopped")
				return
			case <-ticker.C:
				ds.doSyncRepositories(ctx)
				if err := ds.doGarbageCollector(ctx); err != nil {
					log.Printf("[ERROR] %v", err)
				}
			}
		}
	}()
}

func (ds *DataService) doGarbageCollector(ctx context.Context) error {
	ds.mutex.Lock()

	ds.isWorking = true
	defer func() {
		ds.isWorking = false
		ds.mutex.Unlock()
	}()

	if ds.lastSyncDate == 0 {
		return ErrNoSyncedYet
	}

	if err := ds.Storage.RepositoryGarbageCollector(ctx, ds.lastSyncDate); err != nil {
		return fmt.Errorf("repositories garbage collector aborted with error: %v", err)
	}

	if err := ds.Storage.AccessGarbageCollector(ctx); err != nil {
		return fmt.Errorf("access garbage collector aborted with error: %v", err)
	}
	log.Printf("[DEBUG] garbage collector task complete")
	return nil
}
