package service

import (
	"context"
	log "github.com/go-pkgz/lgr"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store/engine"
)

// registryInterface implement method for access data of a registry instance
type registryInterface interface {
	// Catalog return list a set of available repositories in the local registry cluster.
	Catalog(ctx context.Context, n, last string) (registry.Repositories, error)
}

// DataService is service which allow manipulation entries of registry such repositories or tags
type DataService struct {
	Registry registryInterface
	Storage  engine.Interface
}

// SyncExistedRepositories will check existed entries at a registry service and synchronize it
func (ds *DataService) SyncExistedRepositories(ctx context.Context) error {

	var (
		n     string = "20"
		last  string
		total uint64
	)

	go func() {
		for {
			repos, err := ds.Registry.Catalog(context.Background(), n, last)
			total += uint64(len(repos.List))

			if errors.Is(err, registry.ErrNoMorePages) {
				break
			}
			for _, r := range repos.List {
				log.Printf("%s\n", r)
			}
			log.Printf("Total: %d\n", total)

			n, last, err = registry.ParseUrlForNextLink(repos.NextLink)
			if err != nil {
				log.Printf("failed to parse next link: %v", err)
				break
			}

		}
	}()

	return nil
}
