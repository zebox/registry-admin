package service

import (
	"context"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
)

// RepositoryInterface implement methods for manipulation data of repositories
type RepositoryInterface interface {
	CreateRepository(ctx context.Context, entry *store.RegistryEntry) (err error)
	GetRepository(ctx context.Context, entryID int64) (entry store.RegistryEntry, err error)
	FindRepositories(ctx context.Context, filter engine.QueryFilter) (entries engine.ListResponse, err error)
	UpdateRepository(ctx context.Context, conditionClause, data map[string]interface{}) (err error)
	DeleteRepository(ctx context.Context, key string, id interface{})
}

// DataService is service which allow manipulation entries of registry such repositories or tags
type DataService struct {
	repository RepositoryInterface
}
