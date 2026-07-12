// This file implements controlled marketplace download sessions, event
// recording, and download-count snapshot refresh. Download session creation
// verifies release visibility with the download permission before binding
// artifact checksum metadata; catalog and document reads never write events or
// refresh statistics.

package marketplace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	marketv1 "linapro-plugin-marketplace/backend/api/market/v1"
	"linapro-plugin-marketplace/backend/internal/dao"
	"linapro-plugin-marketplace/backend/internal/model/do"
	"linapro-plugin-marketplace/backend/internal/model/entity"
)

const (
	// defaultDownloadSessionTTL is the standard short-lived download window.
	defaultDownloadSessionTTL = 15 * time.Minute
	// maxDownloadSessionTTL caps caller-provided download session lifetimes.
	maxDownloadSessionTTL = 2 * time.Hour
	// marketplaceDownloadSessionIDPrefix identifies marketplace download sessions.
	marketplaceDownloadSessionIDPrefix = "mpdl_"
	// marketplaceDownloadSessionIDRandomBytes controls opaque session ID entropy.
	marketplaceDownloadSessionIDRandomBytes = 16
	// marketplaceDownloadSessionContentPrefix is the controlled content route
	// prefix under the source-plugin API namespace used by the frontend.
	marketplaceDownloadSessionContentPrefix = "/x/linapro-plugin-marketplace/api/v1/market/download-sessions/"
)

// downloadAuthorizationSnapshot is stored with a session for audit visibility.
type downloadAuthorizationSnapshot struct {
	UserID       int64  `json:"userId"`
	TenantID     int64  `json:"tenantId"`
	Permission   string `json:"permission"`
	ArtifactType string `json:"artifactType"`
	GrantedAt    int64  `json:"grantedAt"`
}

// CreateDownloadSession creates a short-lived download authorization session.
func (s *serviceImpl) CreateDownloadSession(
	ctx context.Context,
	in CreateDownloadSessionInput,
) (*DownloadSessionOutput, error) {
	if normalizeKey(in.PluginID) == "" || normalizeKey(in.Version) == "" || in.RequesterUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	subject := downloadVisibilitySubject(in.Visibility, in.RequesterUserID)
	release, err := s.requireVisibleRelease(
		ctx,
		in.PluginID,
		in.Version,
		subject,
		marketplaceVisibilityPermissionDownload,
	)
	if err != nil {
		return nil, err
	}
	artifact, err := s.selectDownloadArtifact(ctx, release, in.ArtifactType)
	if err != nil {
		return nil, err
	}
	if artifact == nil {
		return nil, bizerr.NewCode(CodeMarketplaceDownloadArtifactNotFound)
	}

	sessionID, err := newMarketplaceDownloadSessionID()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	now := time.Now()
	expiresAt := now.Add(normalizeDownloadSessionTTL(in.TTL))
	snapshot, err := downloadAuthorizationSnapshotJSON(subject, artifact, now)
	if err != nil {
		return nil, err
	}
	var sessionIDInt int64
	err = dao.PluginMarketplaceDownloadSession.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		insertID, insertErr := dao.PluginMarketplaceDownloadSession.Ctx(ctx).Data(do.PluginMarketplaceDownloadSession{
			SessionId:             sessionID,
			ReleaseId:             release.Id,
			ArtifactId:            artifact.ID,
			PluginId:              release.PluginId,
			ReleaseVersion:        release.ReleaseVersion,
			RequesterUserId:       in.RequesterUserID,
			Status:                marketv1.MarketplaceDownloadSessionStatusActive.String(),
			ArtifactType:          artifact.ArtifactType.String(),
			ArtifactSizeBytes:     artifact.SizeBytes,
			Sha256:                artifact.Sha256,
			AuthorizationSnapshot: snapshot,
			ExpiresAt:             &expiresAt,
		}).InsertAndGetId()
		if insertErr != nil {
			return bizerr.WrapCode(insertErr, CodeMarketplaceStorageFailed)
		}
		sessionIDInt = insertID
		return s.insertDownloadEvent(ctx, &downloadEventData{
			sessionID:       sessionID,
			releaseID:       release.Id,
			artifactID:      artifact.ID,
			pluginID:        release.PluginId,
			version:         release.ReleaseVersion,
			requesterUserID: in.RequesterUserID,
			eventType:       DownloadEventTypeCreated,
		})
	})
	if err != nil {
		return nil, err
	}

	session, err := s.getDownloadSessionByID(ctx, intID(sessionIDInt))
	if err != nil {
		return nil, err
	}
	return &DownloadSessionOutput{Session: downloadSessionItemFromEntity(session)}, nil
}

// GetDownloadSession returns one requester-owned active or consumed session metadata.
func (s *serviceImpl) GetDownloadSession(
	ctx context.Context,
	in GetDownloadSessionInput,
) (*DownloadSessionOutput, error) {
	session, err := s.requireDownloadSession(ctx, in.SessionID, in.RequesterUserID)
	if err != nil {
		return nil, err
	}
	if err = s.ensureDownloadSessionUsable(ctx, session, in.Visibility); err != nil {
		return nil, err
	}
	return &DownloadSessionOutput{Session: downloadSessionItemFromEntity(session)}, nil
}

// OpenDownloadContent validates a download session and opens the artifact body.
func (s *serviceImpl) OpenDownloadContent(
	ctx context.Context,
	in OpenDownloadContentInput,
) (*OpenDownloadContentOutput, error) {
	session, err := s.requireDownloadSession(ctx, in.SessionID, in.RequesterUserID)
	if err != nil {
		return nil, err
	}
	if err = s.ensureDownloadSessionUsable(ctx, session, in.Visibility); err != nil {
		return nil, err
	}
	if !downloadSessionActiveAndNotExpired(session, time.Now()) {
		return nil, bizerr.NewCode(CodeMarketplaceDownloadSessionUnavailable)
	}
	artifact, err := s.getArtifactRecordByID(ctx, session.ArtifactId)
	if err != nil {
		return nil, err
	}
	if artifact == nil || normalizeKey(artifact.StorageKey) == "" {
		return nil, bizerr.NewCode(CodeMarketplaceDownloadArtifactNotFound)
	}
	if s.artifacts == nil {
		return nil, bizerr.NewCode(CodeMarketplaceStorageFailed)
	}
	body, err := s.artifacts.Open(ctx, artifact.StorageKey)
	if err != nil {
		return nil, err
	}
	if err = s.RecordDownloadEvent(ctx, RecordDownloadEventInput{
		SessionID:       session.SessionId,
		EventType:       DownloadEventTypeStarted,
		RequesterUserID: in.RequesterUserID,
		Visibility:      in.Visibility,
	}); err != nil {
		_ = body.Close()
		return nil, err
	}
	return &OpenDownloadContentOutput{
		Session:  downloadSessionItemFromEntity(session),
		FileName: artifact.FileName,
		Body:     body,
	}, nil
}

// RecordDownloadEvent records one controlled download event for a requester-owned session.
func (s *serviceImpl) RecordDownloadEvent(ctx context.Context, in RecordDownloadEventInput) error {
	session, err := s.requireDownloadSession(ctx, in.SessionID, in.RequesterUserID)
	if err != nil {
		return err
	}
	if err = s.ensureDownloadSessionUsable(ctx, session, in.Visibility); err != nil {
		return err
	}
	eventType := normalizeDownloadEventType(in.EventType)
	if !downloadSessionActiveAndNotExpired(session, time.Now()) {
		return bizerr.NewCode(CodeMarketplaceDownloadSessionUnavailable)
	}
	now := time.Now()
	return dao.PluginMarketplaceDownloadEvent.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if eventType == DownloadEventTypeCompleted {
			cols := dao.PluginMarketplaceDownloadSession.Columns()
			result, updateErr := dao.PluginMarketplaceDownloadSession.Ctx(ctx).
				Where(do.PluginMarketplaceDownloadSession{Id: session.Id}).
				Where(do.PluginMarketplaceDownloadSession{
					Status: marketv1.MarketplaceDownloadSessionStatusActive.String(),
				}).
				Where(cols.ExpiresAt+" > ?", now).
				Data(do.PluginMarketplaceDownloadSession{
					Status:     marketv1.MarketplaceDownloadSessionStatusConsumed.String(),
					ConsumedAt: &now,
				}).
				Update()
			if updateErr != nil {
				return bizerr.WrapCode(updateErr, CodeMarketplaceStorageFailed)
			}
			affected, affectedErr := result.RowsAffected()
			if affectedErr != nil {
				return bizerr.WrapCode(affectedErr, CodeMarketplaceStorageFailed)
			}
			if affected == 0 {
				return bizerr.NewCode(CodeMarketplaceDownloadSessionUnavailable)
			}
		}
		return s.insertDownloadEvent(ctx, &downloadEventData{
			sessionID:       session.SessionId,
			releaseID:       session.ReleaseId,
			artifactID:      session.ArtifactId,
			pluginID:        session.PluginId,
			version:         session.ReleaseVersion,
			requesterUserID: session.RequesterUserId,
			eventType:       eventType,
			clientIPHash:    trimHashBounds(in.ClientIPHash),
			userAgentHash:   trimHashBounds(in.UserAgentHash),
		})
	})
}

// RefreshDownloadStatistics rebuilds the plugin download-count snapshot from completed events.
func (s *serviceImpl) RefreshDownloadStatistics(ctx context.Context, in RefreshDownloadStatisticsInput) error {
	pluginID := normalizeKey(in.PluginID)
	if pluginID == "" {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	total, err := dao.PluginMarketplaceDownloadEvent.Ctx(ctx).
		Where(do.PluginMarketplaceDownloadEvent{
			PluginId:  pluginID,
			EventType: DownloadEventTypeCompleted.String(),
		}).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if _, err = dao.PluginMarketplacePlugin.Ctx(ctx).
		Where(do.PluginMarketplacePlugin{PluginId: pluginID}).
		Data(downloadStatisticsPluginData(total)).
		Update(); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	if _, err = dao.PluginMarketplacePluginReadModel.Ctx(ctx).
		Where(do.PluginMarketplacePluginReadModel{PluginId: pluginID}).
		Data(downloadStatisticsReadModelData(total)).
		Update(); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return nil
}

// downloadEventData carries normalized event fields for persistence.
type downloadEventData struct {
	sessionID       string
	releaseID       int
	artifactID      int
	pluginID        string
	version         string
	requesterUserID int64
	eventType       DownloadEventType
	clientIPHash    string
	userAgentHash   string
}

// insertDownloadEvent writes one immutable download event row.
func (s *serviceImpl) insertDownloadEvent(ctx context.Context, data *downloadEventData) error {
	if data == nil || normalizeKey(data.sessionID) == "" {
		return bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	if _, err := dao.PluginMarketplaceDownloadEvent.Ctx(ctx).Data(do.PluginMarketplaceDownloadEvent{
		SessionId:       data.sessionID,
		ReleaseId:       data.releaseID,
		ArtifactId:      data.artifactID,
		PluginId:        data.pluginID,
		ReleaseVersion:  data.version,
		RequesterUserId: data.requesterUserID,
		EventType:       normalizeDownloadEventType(data.eventType).String(),
		ClientIpHash:    data.clientIPHash,
		UserAgentHash:   data.userAgentHash,
	}).Insert(); err != nil {
		return bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return nil
}

// selectDownloadArtifact returns the requested artifact or the release primary artifact.
func (s *serviceImpl) selectDownloadArtifact(
	ctx context.Context,
	release *entity.PluginMarketplaceRelease,
	artifactType marketv1.MarketplaceArtifactType,
) (*ArtifactRecord, error) {
	if release == nil {
		return nil, bizerr.NewCode(CodeMarketplaceReleaseNotFound)
	}
	if normalizeKey(artifactType.String()) == "" {
		artifacts, err := s.batchPrimaryArtifactsByRelease(ctx, []*entity.PluginMarketplaceRelease{release})
		if err != nil {
			return nil, err
		}
		return artifacts[release.Id], nil
	}
	if !validDownloadArtifactType(artifactType) {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	var row *entity.PluginMarketplaceArtifact
	if err := dao.PluginMarketplaceArtifact.Ctx(ctx).
		Where(do.PluginMarketplaceArtifact{
			ReleaseId:    release.Id,
			ArtifactType: artifactType.String(),
		}).
		Scan(&row); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return artifactRecordFromEntity(row), nil
}

// requireDownloadSession loads one session and verifies requester ownership.
func (s *serviceImpl) requireDownloadSession(
	ctx context.Context,
	sessionID string,
	requesterUserID int64,
) (*entity.PluginMarketplaceDownloadSession, error) {
	if requesterUserID <= 0 {
		return nil, bizerr.NewCode(CodeMarketplaceInvalidInput)
	}
	session, err := s.getDownloadSessionBySessionID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.RequesterUserId != requesterUserID {
		return nil, bizerr.NewCode(CodeMarketplaceDownloadSessionNotFound)
	}
	return session, nil
}

// ensureDownloadSessionUsable verifies status, expiration, and current release visibility.
func (s *serviceImpl) ensureDownloadSessionUsable(
	ctx context.Context,
	session *entity.PluginMarketplaceDownloadSession,
	subject VisibilitySubject,
) error {
	if session == nil {
		return bizerr.NewCode(CodeMarketplaceDownloadSessionNotFound)
	}
	status := marketv1.MarketplaceDownloadSessionStatus(session.Status)
	if !downloadSessionAllowsMetadata(status) {
		if status == marketv1.MarketplaceDownloadSessionStatusExpired {
			return bizerr.NewCode(CodeMarketplaceDownloadSessionExpired)
		}
		return bizerr.NewCode(CodeMarketplaceDownloadSessionUnavailable)
	}
	if session.ExpiresAt == nil || time.Now().After(*session.ExpiresAt) {
		return bizerr.NewCode(CodeMarketplaceDownloadSessionExpired)
	}
	_, err := s.requireVisibleRelease(
		ctx,
		session.PluginId,
		session.ReleaseVersion,
		downloadVisibilitySubject(subject, session.RequesterUserId),
		marketplaceVisibilityPermissionDownload,
	)
	return err
}

// getDownloadSessionBySessionID loads one download session by opaque ID.
func (s *serviceImpl) getDownloadSessionBySessionID(
	ctx context.Context,
	sessionID string,
) (*entity.PluginMarketplaceDownloadSession, error) {
	normalizedID := normalizeKey(sessionID)
	if normalizedID == "" {
		return nil, nil
	}
	var session *entity.PluginMarketplaceDownloadSession
	if err := dao.PluginMarketplaceDownloadSession.Ctx(ctx).
		Where(do.PluginMarketplaceDownloadSession{SessionId: normalizedID}).
		Scan(&session); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return session, nil
}

// getDownloadSessionByID loads one download session by generated ID.
func (s *serviceImpl) getDownloadSessionByID(
	ctx context.Context,
	id int,
) (*entity.PluginMarketplaceDownloadSession, error) {
	if id <= 0 {
		return nil, nil
	}
	var session *entity.PluginMarketplaceDownloadSession
	if err := dao.PluginMarketplaceDownloadSession.Ctx(ctx).
		Where(do.PluginMarketplaceDownloadSession{Id: id}).
		Scan(&session); err != nil {
		return nil, bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return session, nil
}

// downloadSessionItemFromEntity projects one download session to an API DTO.
func downloadSessionItemFromEntity(row *entity.PluginMarketplaceDownloadSession) *marketv1.MarketplaceDownloadSessionItem {
	return downloadSessionItemFromEntityAt(row, time.Now())
}

// downloadSessionItemFromEntityAt projects one download session with a stable clock for tests.
func downloadSessionItemFromEntityAt(row *entity.PluginMarketplaceDownloadSession, now time.Time) *marketv1.MarketplaceDownloadSessionItem {
	if row == nil {
		return nil
	}
	downloadURL := ""
	if downloadSessionActiveAndNotExpired(row, now) {
		downloadURL = marketplaceDownloadSessionContentURL(row.SessionId)
	}
	return &marketv1.MarketplaceDownloadSessionItem{
		SessionId:    row.SessionId,
		PluginId:     row.PluginId,
		Version:      row.ReleaseVersion,
		ArtifactType: marketv1.MarketplaceArtifactType(row.ArtifactType),
		SizeBytes:    row.ArtifactSizeBytes,
		Sha256:       row.Sha256,
		Status:       marketv1.MarketplaceDownloadSessionStatus(row.Status),
		DownloadUrl:  downloadURL,
		ExpiresAt:    unixMillisPtr(row.ExpiresAt),
		ConsumedAt:   unixMillisPtr(row.ConsumedAt),
		CreatedAt:    unixMillisPtr(row.CreatedAt),
	}
}

// downloadVisibilitySubject fills the user scope from the requester when needed.
func downloadVisibilitySubject(subject VisibilitySubject, requesterUserID int64) VisibilitySubject {
	if subject.UserID <= 0 && requesterUserID > 0 {
		subject.UserID = requesterUserID
	}
	return subject
}

// downloadStatisticsPluginData builds the authoritative plugin snapshot update.
func downloadStatisticsPluginData(total int) do.PluginMarketplacePlugin {
	return do.PluginMarketplacePlugin{DownloadCount: total}
}

// downloadStatisticsReadModelData builds the catalog read-model snapshot update.
func downloadStatisticsReadModelData(total int) do.PluginMarketplacePluginReadModel {
	return do.PluginMarketplacePluginReadModel{DownloadCount: total}
}

// normalizeDownloadSessionTTL applies default and maximum download lifetimes.
func normalizeDownloadSessionTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return defaultDownloadSessionTTL
	}
	if ttl > maxDownloadSessionTTL {
		return maxDownloadSessionTTL
	}
	return ttl
}

// normalizeDownloadEventType defaults unknown event types to started.
func normalizeDownloadEventType(eventType DownloadEventType) DownloadEventType {
	switch eventType {
	case DownloadEventTypeCreated, DownloadEventTypeStarted, DownloadEventTypeCompleted, DownloadEventTypeFailed:
		return eventType
	default:
		return DownloadEventTypeStarted
	}
}

// validDownloadArtifactType reports whether a caller-provided artifact type is supported.
func validDownloadArtifactType(artifactType marketv1.MarketplaceArtifactType) bool {
	switch artifactType {
	case marketv1.MarketplaceArtifactTypeSourceZip,
		marketv1.MarketplaceArtifactTypeDynamicZip,
		marketv1.MarketplaceArtifactTypePluginWasm:
		return true
	default:
		return false
	}
}

// downloadSessionAllowsMetadata reports whether session metadata can still be read.
func downloadSessionAllowsMetadata(status marketv1.MarketplaceDownloadSessionStatus) bool {
	switch status {
	case marketv1.MarketplaceDownloadSessionStatusActive,
		marketv1.MarketplaceDownloadSessionStatusConsumed:
		return true
	default:
		return false
	}
}

// marketplaceDownloadSessionContentURL returns the controlled artifact content route.
func marketplaceDownloadSessionContentURL(sessionID string) string {
	normalizedID := normalizeKey(sessionID)
	if normalizedID == "" {
		return ""
	}
	return marketplaceDownloadSessionContentPrefix + normalizedID + "/content"
}

// newMarketplaceDownloadSessionID creates an opaque URL-safe session ID.
func newMarketplaceDownloadSessionID() (string, error) {
	randomBytes := make([]byte, marketplaceDownloadSessionIDRandomBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return marketplaceDownloadSessionIDPrefix + hex.EncodeToString(randomBytes), nil
}

// downloadAuthorizationSnapshotJSON serializes the authorization decision snapshot.
func downloadAuthorizationSnapshotJSON(
	subject VisibilitySubject,
	artifact *ArtifactRecord,
	grantedAt time.Time,
) (string, error) {
	if artifact == nil {
		return "", bizerr.NewCode(CodeMarketplaceDownloadArtifactNotFound)
	}
	content, err := json.Marshal(&downloadAuthorizationSnapshot{
		UserID:       subject.UserID,
		TenantID:     subject.TenantID,
		Permission:   string(marketplaceVisibilityPermissionDownload),
		ArtifactType: artifact.ArtifactType.String(),
		GrantedAt:    grantedAt.UnixMilli(),
	})
	if err != nil {
		return "", bizerr.WrapCode(err, CodeMarketplaceStorageFailed)
	}
	return string(content), nil
}

// downloadSessionActiveAndNotExpired reports whether a session can stream content.
func downloadSessionActiveAndNotExpired(session *entity.PluginMarketplaceDownloadSession, now time.Time) bool {
	if session == nil || marketv1.MarketplaceDownloadSessionStatus(session.Status) != marketv1.MarketplaceDownloadSessionStatusActive {
		return false
	}
	return session.ExpiresAt != nil && now.Before(*session.ExpiresAt)
}

// trimHashBounds keeps optional event hash fields within table limits.
func trimHashBounds(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) > 128 {
		return trimmed[:128]
	}
	return trimmed
}
