package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/notifications"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/store"
	"github.com/zebox/registry-admin/app/store/engine"
)

func (ds *DataService) RepositoryEventsProcessing(ctx context.Context, envelope notifications.Envelope) (err error) {

	for _, e := range envelope.Events {
		switch e.Action {
		case notifications.EventActionPush, notifications.EventActionPull:
			if errUpdate := ds.updateRepositoryEntry(ctx, e); errUpdate != nil {
				err = multierror.Append(err, errUpdate)
			}
			return err
		case notifications.EventActionDelete:
			return ds.deleteRepositoryEntry(ctx, e)
		}
	}
	return errors.New("unsupported event")
}

// updateRepositoryEntry will create repository entry if it doesn't exist or update when already exist
func (ds *DataService) updateRepositoryEntry(ctx context.Context, event notifications.Event) error {
	filter := engine.QueryFilter{
		Filters: map[string]interface{}{"repository_name": event.Target.Repository, "tag": event.Target.Tag},
	}

	result, err := ds.Storage.FindRepositories(ctx, filter)
	if err != nil {
		return err
	}

	// increase pull counter when repo pull
	if event.Action == notifications.EventActionPull && result.Total == 1 {
		repositoryEntry := result.Data[0].(store.RegistryEntry)

		err = ds.Storage.UpdateRepository(
			ctx,
			map[string]interface{}{"id": repositoryEntry.ID},                        // condition
			map[string]interface{}{"pull_counter": repositoryEntry.PullCounter + 1}, // data for update
		)
		return err
	}

	eventRawBytes, errJSON := json.Marshal(event)
	if errJSON != nil {
		return errors.Wrap(errJSON, "failed to marshalling event raw data")
	}

	if result.Total == 0 {
		digest := ""
		for _, ref := range event.Target.References {
			if ref.MediaType == schema2.MediaTypeImageConfig {
				digest = ref.Digest.String()
				break
			}
		}
		if digest == "" {
			return fmt.Errorf("digest not found for repo: %s and tag %s", event.Target.Repository, event.Target.Tag)
		}

		repositoryEntry := &store.RegistryEntry{
			RepositoryName: event.Target.Repository,
			Tag:            event.Target.Tag,
			Digest:         digest,
			Size:           event.Target.Size,
			Timestamp:      event.Timestamp.Unix(),
			Raw:            string(eventRawBytes),
		}
		err = ds.Storage.CreateRepository(ctx, repositoryEntry)
		return err
	}

	if result.Total == 1 {
		repositoryEntry := result.Data[0].(store.RegistryEntry)

		err = ds.Storage.UpdateRepository(
			ctx,
			map[string]interface{}{"id": repositoryEntry.ID},                                                             // condition
			map[string]interface{}{"size": event.Target.Size, "timestamp": event.Timestamp.Unix(), "raw": eventRawBytes}, // data for update
		)
		return err
	}

	return errors.Errorf("query filter returned multiple result: %v+", filter.Filters)
}

// deleteRepositoryEntry deletes repository entry by an event delete
func (ds *DataService) deleteRepositoryEntry(ctx context.Context, event notifications.Event) error {

	digest := event.Target.Descriptor.Digest

	if digest == "" {
		return errors.Errorf("failed to delete image from repository %s, entry %s not found", event.Target.Repository, digest)
	}

	if err := ds.Storage.DeleteRepository(ctx, "digest", digest); err != nil && err != engine.ErrNotFound {
		return errors.Errorf("failed to delete image entry digest: %s err: %v", digest, err)
	}
	return nil
}
