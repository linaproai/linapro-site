// This file implements publisher profile creation and lookup helpers for the
// marketplace service.

package marketplace

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

// ListPublishers returns publisher profiles available to the current operator.
func (s *serviceImpl) ListPublishers(ctx context.Context, in ListPublishersInput) (*PublisherListOutput, error) {
	pageNum, pageSize := normalizeMarketplacePage(in.PageNum, in.PageSize)
	model := dao.PluginMarketplacePublisher.Ctx(ctx)
	if in.OwnerUserID > 0 {
		model = applyPublisherOwnerFilter(model, in.OwnerUserID)
	}
	if keyword := normalizeKey(in.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		model = model.Where(
			"(publisher_key LIKE ? OR name LIKE ? OR summary LIKE ?)",
			like,
			like,
			like,
		)
	}
	total, err := model.Clone().Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	var rows []*entity.PluginMarketplacePublisher
	cols := dao.PluginMarketplacePublisher.Columns()
	if err = model.Clone().
		OrderDesc(cols.UpdatedAt).
		OrderDesc(cols.Id).
		Page(pageNum, pageSize).
		Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	items := make([]*marketv1.MarketplacePublisherItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, publisherItemFromEntity(row, "", false))
	}
	return &PublisherListOutput{List: items, Total: total}, nil
}

// CreatePublisher creates a marketplace publisher profile for a publishing owner.
// Each owner user may bind at most one publisher profile.
func (s *serviceImpl) CreatePublisher(ctx context.Context, in CreatePublisherInput) (*PublisherRecord, error) {
	normalizedKey := normalizeKey(in.PublisherKey)
	if normalizedKey == "" || normalizeKey(in.Name) == "" || in.OwnerUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}

	ownedCount, err := applyPublisherOwnerFilter(
		dao.PluginMarketplacePublisher.Ctx(ctx),
		in.OwnerUserID,
	).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if ownedCount > 0 {
		return nil, bizerr.NewCode(CodeMarketplacePublisherOwnerAlreadyBound)
	}

	existing, err := s.getPublisherByKey(ctx, normalizedKey)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, bizerr.NewCode(CodeMarketplacePublisherAlreadyExists)
	}

	id, err := dao.PluginMarketplacePublisher.Ctx(ctx).Data(do.PluginMarketplacePublisher{
		PublisherKey: normalizedKey,
		Name:         normalizeKey(in.Name),
		Summary:      normalizeKey(in.Summary),
		OwnerUserId:  in.OwnerUserID,
		OwnerOrgId:   in.OwnerOrgID,
		Verified:     false,
		Status:       PublisherStatusActive.String(),
		Homepage:     normalizeKey(in.Homepage),
		ContactEmail: normalizeKey(in.ContactEmail),
	}).InsertAndGetId()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return s.getPublisherByID(ctx, intID(id))
}

// UpdatePublisher updates key and mutable fields of a publisher profile owned by the operator.
func (s *serviceImpl) UpdatePublisher(ctx context.Context, in UpdatePublisherInput) (*PublisherRecord, error) {
	currentKey := normalizeKey(in.CurrentPublisherKey)
	nextKey := normalizeKey(in.PublisherKey)
	if currentKey == "" || nextKey == "" || normalizeKey(in.Name) == "" || in.OwnerUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	publisher, err := s.requirePublisherOwnedByUser(ctx, currentKey, in.OwnerUserID)
	if err != nil {
		return nil, err
	}

	if nextKey != publisher.PublisherKey {
		existing, lookupErr := s.getPublisherByKey(ctx, nextKey)
		if lookupErr != nil {
			return nil, lookupErr
		}
		if existing != nil && existing.Id != publisher.Id {
			return nil, bizerr.NewCode(CodeMarketplacePublisherAlreadyExists)
		}
	}

	if _, err = dao.PluginMarketplacePublisher.Ctx(ctx).
		Where(dao.PluginMarketplacePublisher.Columns().Id, publisher.Id).
		Data(do.PluginMarketplacePublisher{
			PublisherKey: nextKey,
			Name:         normalizeKey(in.Name),
			Summary:      normalizeKey(in.Summary),
			Homepage:     normalizeKey(in.Homepage),
			ContactEmail: normalizeKey(in.ContactEmail),
		}).
		Update(); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return s.getPublisherByID(ctx, publisher.Id)
}

// requirePublisherOwnedByUser returns an active publisher owned by one user.
func (s *serviceImpl) requirePublisherOwnedByUser(
	ctx context.Context,
	publisherKey string,
	ownerUserID int64,
) (*entity.PluginMarketplacePublisher, error) {
	if ownerUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	publisher, err := s.getPublisherByKeyForOwner(ctx, publisherKey, ownerUserID)
	if err != nil {
		return nil, err
	}
	if publisher == nil {
		return nil, bizerr.NewCode(CodeMarketplacePublisherNotFound)
	}
	if PublisherStatus(publisher.Status) != PublisherStatusActive {
		return nil, bizerr.NewCode(CodeMarketplacePublisherUnavailable)
	}
	return publisher, nil
}

// requirePublisherIDOwnedByUser returns an active publisher ID owned by one user.
func (s *serviceImpl) requirePublisherIDOwnedByUser(
	ctx context.Context,
	publisherID int,
	ownerUserID int64,
) (*entity.PluginMarketplacePublisher, error) {
	if publisherID <= 0 || ownerUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	var publisher *entity.PluginMarketplacePublisher
	if err := applyPublisherOwnerFilter(
		dao.PluginMarketplacePublisher.Ctx(ctx).
			Where(dao.PluginMarketplacePublisher.Columns().Id, publisherID),
		ownerUserID,
	).Scan(&publisher); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if publisher == nil {
		return nil, bizerr.NewCode(CodeMarketplacePublisherNotFound)
	}
	if PublisherStatus(publisher.Status) != PublisherStatusActive {
		return nil, bizerr.NewCode(CodeMarketplacePublisherUnavailable)
	}
	return publisher, nil
}

// applyPublisherOwnerFilter restricts a publisher query to one owning user.
func applyPublisherOwnerFilter(model *gdb.Model, ownerUserID int64) *gdb.Model {
	return model.Where(dao.PluginMarketplacePublisher.Columns().OwnerUserId, ownerUserID)
}

// getPublisherByKey loads one publisher by stable key.
func (s *serviceImpl) getPublisherByKey(ctx context.Context, publisherKey string) (*entity.PluginMarketplacePublisher, error) {
	normalizedKey := normalizeKey(publisherKey)
	if normalizedKey == "" {
		return nil, nil
	}

	var publisher *entity.PluginMarketplacePublisher
	if err := dao.PluginMarketplacePublisher.Ctx(ctx).
		Where(do.PluginMarketplacePublisher{PublisherKey: normalizedKey}).
		Scan(&publisher); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return publisher, nil
}

// getPublisherByKeyForOwner loads one publisher by stable key and owning user.
func (s *serviceImpl) getPublisherByKeyForOwner(
	ctx context.Context,
	publisherKey string,
	ownerUserID int64,
) (*entity.PluginMarketplacePublisher, error) {
	normalizedKey := normalizeKey(publisherKey)
	if normalizedKey == "" {
		return nil, nil
	}

	var publisher *entity.PluginMarketplacePublisher
	if err := applyPublisherOwnerFilter(
		dao.PluginMarketplacePublisher.Ctx(ctx).
			Where(dao.PluginMarketplacePublisher.Columns().PublisherKey, normalizedKey),
		ownerUserID,
	).Scan(&publisher); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return publisher, nil
}

// getPublisherByID loads one publisher by primary key and projects it to a service record.
func (s *serviceImpl) getPublisherByID(ctx context.Context, id int) (*PublisherRecord, error) {
	var publisher *entity.PluginMarketplacePublisher
	if err := dao.PluginMarketplacePublisher.Ctx(ctx).
		Where(do.PluginMarketplacePublisher{Id: id}).
		Scan(&publisher); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if publisher == nil {
		return nil, bizerr.NewCode(CodeMarketplacePublisherNotFound)
	}
	return publisherRecordFromEntity(publisher), nil
}
