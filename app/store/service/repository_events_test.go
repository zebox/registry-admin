package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
	"testing"
	"time"
)

func TestDataService_RepositoryEventsProcessing(t *testing.T) {

}

func prepareEngineMock() *engine.InterfaceMock {
	testRepositoriesEntries := []store.RegistryEntry{
		{
			RepositoryName: "test/repo_1",
			Tag:            "1.1.0",
		},
		{
			RepositoryName: "test/repo_1",
			Tag:            "1.2.0",
		},
		{
			RepositoryName: "test/repo_2",
			Tag:            "2.1.0",
		},
		{
			RepositoryName: "test/repo_2",
			Tag:            "2.2.0",
		},
	}

	return &engine.InterfaceMock{
		CreateRepositoryFunc: func(ctx context.Context, entry *store.RegistryEntry) error {
			for _, testEntry := range testRepositoriesEntries {
				if testEntry.RepositoryName == entry.RepositoryName && testEntry.Tag == entry.Tag {
					return errors.New(" duplicate repository entry")
				}

				hasher := sha256.New()
				hasher.Write([]byte(entry.RepositoryName + entry.Tag))
				sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
				entry.Digest = "sha256:" + sha
				entry.Timestamp = time.Now().Unix()

			}
			return nil
		},

		UpdateRepositoryFunc: func(ctx context.Context, conditionClause map[string]interface{}, data map[string]interface{}) error {
			var id int64
			if v, ok := conditionClause["id"]; ok {
				id = v.(int64)
			}
			for _, testEntry := range testRepositoriesEntries {
				if id == testEntry.ID {
					return nil
				}
			}
			return errors.New("entry not found")
		},
	}
}
