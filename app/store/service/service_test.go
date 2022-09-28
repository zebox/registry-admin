package service

import (
	"context"
	"github.com/zebox/registry-admin/app/registry"
	"testing"
)

func prepareRegistryMock(t *testing.T) registryInterfaceMock {
	return registryInterfaceMock{
		CatalogFunc: func(ctx context.Context, n string, last string) (registry.Repositories, error) {
			return registry.Repositories{}, nil
		},
	}
}
