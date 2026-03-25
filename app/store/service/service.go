package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"

	log "github.com/go-pkgz/lgr"
)

const defaultPageSize = "50" // the number of repository items for pagination request when catalog listing

var errNoSyncedYet = errors.New("garbage collector skip because sync required start first")

// registryInterface implement method for access data of a registry instance
type registryInterface interface {
	// Catalog return list a set of available repositories in the local registry cluster.
	Catalog(ctx context.Context, n, last string) (registry.Repositories, error)

	// ListingImageTags retrieve information about tags.
	ListingImageTags(ctx context.Context, repoName, n, last string) (registry.ImageTags, error)

	// Manifest will fetch the manifest identified by 'name' and 'reference' where 'reference' can be a tag or digest.
	Manifest(ctx context.Context, repoName, tag string, manifest registry.ManifestListSchemaItem) (registry.ManifestSchemaV2, error)

	// ManifestList will fetch the manifest identified by 'name' and 'reference' where 'reference' can be a tag or digest.
	ManifestList(ctx context.Context, repoName, tag string) (registry.ManifestListSchema, error)
}

// DataService is service which allow manipulation entries of registry such repositories or tags
type DataService struct {
	Registry registryInterface
	Storage  engine.Interface

	lastSyncDate atomic.Value // timestamp for mark actual record and use it for repository garbage collector

	// used for checks status either syncing operation or garbage collector is active
	// it prevents stacking task in queue and run in parallels
	isWorking atomic.Value

	syncGcChan chan context.Context
}

// SyncExistedRepositories will check existed entries at a registry service and synchronize it
func (ds *DataService) SyncExistedRepositories(ctx context.Context) error {

	if ds.isWorking.Load().(bool) {
		return errors.New("repository sync currently running")
	}
	ds.syncGcChan <- ctx
	return nil
}

func (ds *DataService) doSyncRepositories(ctx context.Context) {

	now := time.Now().Unix()

	var (
		n          = defaultPageSize // item number per page
		lastRepo   string
		totalRepos uint64
		lastTag    string
		totalTags  uint64
	)
	for {
		repos, errCatalog := ds.Registry.Catalog(ctx, n, lastRepo)

		if errCatalog != nil && errCatalog != registry.ErrNoMorePages {
			log.Printf("[ERROR] failed to fetch catalog list: %v", errCatalog)
			log.Printf("[ERROR] break on page with last repository: %s", lastRepo)
			break
		}

		totalRepos += uint64(len(repos.List))

		for _, repo := range repos.List {

			log.Printf("[DEBUG] Repository name: %s\n", repo)

			for {
				tags, errTags := ds.Registry.ListingImageTags(ctx, repo, defaultPageSize, lastTag)
				if errTags != nil && !errors.Is(errTags, registry.ErrNoMorePages) {
					log.Printf("[ERROR] failed to get repository tags at repository '%s': %v", repo, errTags)
					log.Printf("[ERROR] skip this repository: %s", repo)
					tags = registry.ImageTags{
						Name: repo,
						Tags: []string{
							"-",
						},
						NextLink: "",
					}
				}

				totalTags += uint64(len(tags.Tags))

				if tags.Tags != nil {
					for _, tag := range tags.Tags {
						log.Printf("[DEBUG] Tag name: %s\n", tag)
						manifestList, errManifestList := ds.Registry.ManifestList(ctx, repo, tag)
						if errManifestList != nil {
							log.Printf("[ERROR] failed to fetch manifest-list from repo '%s' for tag '%s' errCatalog: %s", repo, tag, errManifestList)
							log.Printf("[ERROR] ManifestList skip this tag: %s:%s", repo, tag)
							manifestList = registry.ManifestListSchema{
								SchemaVersion: 0,
								MediaType:     "not_exists_type",
								Manifests: []registry.ManifestListSchemaItem{
									{
										MediaType: "not_exists_type",
										Size:      0,
										Digest:    tag,
										Platform: registry.ManifestPlatform{
											OS:           "unknown",
											Architecture: "unknown",
										},
									},
								},
								TotalSize:     0,
								ContentDigest: "",
							}
						}
						// log.Printf("[ERROR] %v", manifestList)
						for _, m := range manifestList.Manifests {
							log.Printf("[DEBUG] Manifest: %v\n", m)
							var manifest registry.ManifestSchemaV2
							var errManifest error
							if m.MediaType == "not_exists_type" {
								errManifest = errors.New("manifest list not exists")
							} else {
								manifest, errManifest = ds.Registry.Manifest(ctx, repo, m.Digest, m)
							}
							var rawManifestData []byte
							if errManifest != nil {
								log.Printf("[ERROR] failed to fetch manifest from repo '%s' for tag '%s' errCatalog: %s", repo, tag, errManifest)
								log.Printf("[ERROR] Manifest: skip this tag: %s:%s", repo, tag)
								manifest.MediaType = m.MediaType
								manifest.ContentDigest = m.Digest
								manifest.ConfigDescriptor.Digest = m.Digest
								manifest.TotalSize = m.Size
								manifest.Platform.OS = m.Platform.OS
								manifest.Platform.Architecture = m.Platform.Architecture
								rawManifestData = []byte(fmt.Sprintf(`{"error": "%s", "tag": "%s", "os": "%s", "arch": "%s", "digest": "%s"}`, errManifest, tag, m.Platform.OS, m.Platform.Architecture, m.Digest))
							} else {
								log.Printf("[DEBUG] Manifest: %v\n", manifest)
								var errMarshal error
								rawManifestData, errMarshal = json.Marshal(&manifest)
								if errMarshal != nil {
									log.Printf("[ERROR] failed to marshal manifest data from repo '%s' for tag '%s' errCatalog: %s", repo, tag, errMarshal)
								}
							}
							manifest.Digest = manifest.ConfigDescriptor.Digest
							manifest.ContentDigest = m.Digest
							manifest.Platform.OS = m.Platform.OS
							manifest.Platform.Architecture = m.Platform.Architecture

							entry := &store.RegistryEntry{
								RepositoryName: repo,
								// Tag:            fmt.Sprintf("%s:%s:%s:%s:%s", tag, m.Platform.OS, m.Platform.Architecture, m.Digest, manifest.Digest),
								Tag:          fmt.Sprintf("%s:%s:%s", tag, m.Platform.OS, m.Platform.Architecture),
								Digest:       manifest.ContentDigest,
								ConfigDigest: manifest.ConfigDescriptor.Digest,
								Size:         manifest.TotalSize,
								Timestamp:    now,
								Raw:          string(rawManifestData),
							}
							if errCreate := ds.Storage.CreateRepository(ctx, entry); errCreate != nil {
								if !strings.HasPrefix(errCreate.Error(), "UNIQUE") {
									log.Printf("[ERROR] errCreate: failed to marshal manifest data from repo '%s' for tag '%s' err: %s", repo, tag, errCreate)
									log.Printf("[ERROR] errCreate: skip this tag: %s:%s", repo, tag)
									continue
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
									if errUpdate.Error() != "record didn't update" {
										log.Printf("[ERROR] errUpdate: storage update failed for repo '%s' and tag '%s' with error: %v, error on create: %v", repo, tag, errUpdate, errCreate)
									}
									continue
								}
								continue
							}
							log.Printf("[DEBUG] New entry added: repo: '%s', tag: '%s'", repo, tag)
						}
					}
				}

				if errors.Is(errTags, registry.ErrNoMorePages) {
					log.Printf("[DEBUG] Tags for '%s' synced. Total: %d\n", repo, totalTags)
					lastTag = ""
					break
				}

				_, lastTag, errTags = registry.ParseURLForNextLink(tags.NextLink)
				if errTags != nil {
					log.Printf("[ERROR] failed to parse next link: %v", errTags)
					log.Printf("[ERROR] skip this repository: %s", repo)
					continue
				}
			}
		}

		if errors.Is(errCatalog, registry.ErrNoMorePages) {
			break
		}

		n, lastRepo, errCatalog = registry.ParseURLForNextLink(repos.NextLink)
		if errCatalog != nil {
			log.Printf("[ERROR] failed to parse next link: %v", errCatalog)
			break
		}
	}

	if totalTags > 0 {
		ds.lastSyncDate.Store(now)
		log.Printf("[INFO] Repositories synced. Total: %d\n", totalRepos)
	}
}

// RepositoriesMaintenance check repositories for outdated or updated data in repository storage
// with 'lastSyncDate' value. Timestamp field update at every sync call in repository storage
// and compare with 'lastSyncDate' variable.
// If values above is different garbage collector will remove all outdated entries
func (ds *DataService) RepositoriesMaintenance(ctx context.Context, timeout int64) {

	if timeout == 0 {
		timeout = 60
	}
	ticker := time.NewTicker(time.Duration(timeout) * time.Minute)
	ds.isWorking.Store(false)
	ds.lastSyncDate.Store(int64(0))
	ds.syncGcChan = make(chan context.Context)

	SyncGcTaskFn := func(syncCtx context.Context) {
		if ds.isWorking.Load().(bool) {
			return
		}

		ds.isWorking.Store(true)
		defer func() {
			ds.isWorking.Store(false)
		}()

		ds.doSyncRepositories(syncCtx)
		if err := ds.doGarbageCollector(syncCtx); err != nil {
			log.Printf("[ERROR] %v", err)
		}
	}

	// starting garbage collector background task
	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Printf("[DEBUG] repositories maintaining task stopped")
				return
			case <-ticker.C:
				SyncGcTaskFn(ctx)
			case handleCtx := <-ds.syncGcChan: // handle manual start sync and garbage collector
				SyncGcTaskFn(handleCtx)
			}
		}
	}()
}

func (ds *DataService) doGarbageCollector(ctx context.Context) error {

	lastSyncDate := ds.lastSyncDate.Load().(int64)
	if lastSyncDate == 0 {
		return errNoSyncedYet
	}

	if err := ds.Storage.RepositoryGarbageCollector(ctx, lastSyncDate); err != nil {
		return fmt.Errorf("repositories garbage collector aborted with error: %v", err)
	}

	if err := ds.Storage.AccessGarbageCollector(ctx); err != nil {
		return fmt.Errorf("access garbage collector aborted with error: %v", err)
	}

	log.Printf("[DEBUG] garbage collector task complete")
	return nil
}
