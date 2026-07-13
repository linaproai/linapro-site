// This file implements marketplace plugin identity and ownership operations.

package marketplace

import (
	"context"

	"lina-core/pkg/bizerr"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

// SavePluginDraft creates or updates a plugin identity owned by one publisher.
func (s *serviceImpl) SavePluginDraft(ctx context.Context, in SavePluginDraftInput) (*PluginRecord, error) {
	publisher, err := s.requirePublisherOwnedByUser(ctx, in.PublisherKey, in.OwnerUserID)
	if err != nil {
		return nil, err
	}

	pluginID := normalizeKey(in.PluginID)
	if pluginID == "" || normalizeKey(in.Name) == "" || normalizeKey(in.Summary) == "" {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}

	existing, err := s.getPluginByID(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.PublisherId != publisher.Id {
			return nil, bizerr.NewCode(CodeMarketplacePluginIDOwned)
		}
		return s.updatePluginDraft(ctx, existing.Id, in)
	}

	id, err := dao.PluginMarketplacePlugin.Ctx(ctx).Data(do.PluginMarketplacePlugin{
		PublisherId:     publisher.Id,
		PluginId:        pluginID,
		Name:            normalizeKey(in.Name),
		Summary:         normalizeKey(in.Summary),
		Description:     normalizeKey(in.Description),
		PluginType:      normalizePluginType(in.PluginType).String(),
		MarketStatus:    marketv1.MarketplaceStatusDraft.String(),
		Visibility:      normalizeVisibility(in.Visibility).String(),
		LatestReleaseId: 0,
		LatestVersion:   "",
		Icon:            normalizeKey(in.Icon),
		Homepage:        normalizeKey(in.Homepage),
		Repository:      normalizeKey(in.Repository),
		License:         normalizeKey(in.License),
		DownloadCount:   0,
		SourceKind:      uploadSourceKind,
	}).InsertAndGetId()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return s.getPluginRecordByID(ctx, intID(id))
}

// updatePluginDraft updates owner-controlled plugin identity metadata.
func (s *serviceImpl) updatePluginDraft(ctx context.Context, id int, in SavePluginDraftInput) (*PluginRecord, error) {
	data := do.PluginMarketplacePlugin{
		Name:        normalizeKey(in.Name),
		Summary:     normalizeKey(in.Summary),
		Description: normalizeKey(in.Description),
		PluginType:  normalizePluginType(in.PluginType).String(),
		Visibility:  normalizeVisibility(in.Visibility).String(),
		Icon:        normalizeKey(in.Icon),
		Homepage:    normalizeKey(in.Homepage),
		Repository:  normalizeKey(in.Repository),
		License:     normalizeKey(in.License),
	}
	if _, err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{Id: id}).
		Data(data).
		Update(); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return s.getPluginRecordByID(ctx, id)
}

// requireOwnedPlugin returns a plugin only when the publisher owns its plugin ID.
func (s *serviceImpl) requireOwnedPlugin(
	ctx context.Context,
	publisher *entity.PluginMarketplacePublisher,
	pluginID string,
) (*entity.PluginMarketplacePlugin, error) {
	if publisher == nil {
		return nil, bizerr.NewCode(CodeMarketplacePublisherNotFound)
	}
	plugin, err := s.getPluginByID(ctx, pluginID)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	if plugin.PublisherId != publisher.Id {
		return nil, bizerr.NewCode(CodeMarketplacePluginIDOwned)
	}
	return plugin, nil
}

// requirePluginForPublisher resolves a plugin by optional publisher key. When the
// publisher key is empty, the plugin's existing owner publisher is used.
func (s *serviceImpl) requirePluginForPublisher(
	ctx context.Context,
	publisherKey string,
	pluginID string,
	ownerUserID int64,
) (*entity.PluginMarketplacePlugin, error) {
	if ownerUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	if normalizeKey(publisherKey) != "" {
		publisher, err := s.requirePublisherOwnedByUser(ctx, publisherKey, ownerUserID)
		if err != nil {
			return nil, err
		}
		return s.requireOwnedPlugin(ctx, publisher, pluginID)
	}
	plugin, err := s.getPluginByIDForOwner(ctx, pluginID, ownerUserID)
	if err != nil {
		return nil, err
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	if _, err = s.requirePublisherIDOwnedByUser(ctx, plugin.PublisherId, ownerUserID); err != nil {
		return nil, err
	}
	return plugin, nil
}

// resolvePublisherKeyForPlugin returns the owned publisher key for one plugin identity.
func (s *serviceImpl) resolvePublisherKeyForPlugin(
	ctx context.Context,
	publisherKey string,
	pluginID string,
	ownerUserID int64,
) (string, error) {
	plugin, err := s.requirePluginForPublisher(ctx, publisherKey, pluginID, ownerUserID)
	if err != nil {
		return "", err
	}
	publisher, err := s.requirePublisherIDOwnedByUser(ctx, plugin.PublisherId, ownerUserID)
	if err != nil {
		return "", err
	}
	return publisher.PublisherKey, nil
}

// getPluginByIDForOwner loads one marketplace plugin through its owner publisher.
func (s *serviceImpl) getPluginByIDForOwner(
	ctx context.Context,
	pluginID string,
	ownerUserID int64,
) (*entity.PluginMarketplacePlugin, error) {
	normalizedID := normalizeKey(pluginID)
	if normalizedID == "" || ownerUserID <= 0 {
		return nil, nil
	}

	var plugin *entity.PluginMarketplacePlugin
	pluginCols := dao.PluginMarketplacePlugin.Columns()
	if err := applyOwnerPublisherFilter(
		dao.PluginMarketplacePlugin.Ctx(ctx).
			Where(pluginCols.PluginId, normalizedID),
		dao.PluginMarketplacePublisher.Ctx(ctx),
		ownerUserID,
	).Scan(&plugin); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return plugin, nil
}

// getPluginByID loads one marketplace plugin by stable plugin ID.
func (s *serviceImpl) getPluginByID(ctx context.Context, pluginID string) (*entity.PluginMarketplacePlugin, error) {
	normalizedID := normalizeKey(pluginID)
	if normalizedID == "" {
		return nil, nil
	}

	var plugin *entity.PluginMarketplacePlugin
	if err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{PluginId: normalizedID}).
		Scan(&plugin); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return plugin, nil
}

// getPluginRecordByID loads one marketplace plugin by primary key and projects it to a service record.
func (s *serviceImpl) getPluginRecordByID(ctx context.Context, id int) (*PluginRecord, error) {
	var plugin *entity.PluginMarketplacePlugin
	if err := dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{Id: id}).
		Scan(&plugin); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if plugin == nil {
		return nil, bizerr.NewCode(CodeMarketplacePluginNotFound)
	}
	return pluginRecordFromEntity(plugin), nil
}
